#!/usr/bin/env python3
"""Compare refs line scanning with a derived identifier-per-line index."""

from __future__ import annotations

import argparse
import json
import os
import statistics
import subprocess
import tempfile
import time
import unicodedata
from pathlib import Path


DEFAULT_QUERY = "assert_valid_keys"
DEFAULT_UPDATE_PATH = "activesupport/lib/active_support/core_ext/hash/keys.rb"


def run(command: list[str], *, capture: bool = False) -> str:
    result = subprocess.run(
        command,
        check=True,
        text=True,
        stdout=subprocess.PIPE if capture else subprocess.DEVNULL,
        stderr=subprocess.PIPE,
    )
    return result.stdout if capture else ""


def timed(command: list[str], *, capture: bool = False) -> tuple[int, str]:
    started = time.perf_counter_ns()
    output = run(command, capture=capture)
    return time.perf_counter_ns() - started, output


def median_command(command: list[str], repeats: int) -> tuple[int, str]:
    durations = []
    output = ""
    for _ in range(repeats):
        duration, output = timed(command, capture=True)
        durations.append(duration)
    return int(statistics.median(durations)), output


def is_identifier_char(char: str) -> bool:
    if char in "_$":
        return True
    category = unicodedata.category(char)
    return category.startswith("L") or category == "Nd"


def identifiers(text: str) -> set[str]:
    """Return distinct Go-compatible identifier runs on a source line."""
    found: set[str] = set()
    start = None
    for offset, char in enumerate(text):
        if is_identifier_char(char):
            if start is None:
                start = offset
        elif start is not None:
            found.add(text[start:offset])
            start = None
    if start is not None:
        found.add(text[start:])
    return found


def materialize_index(db_path: Path) -> int:
    import sqlite3

    started = time.perf_counter_ns()
    connection = sqlite3.connect(db_path)
    try:
        connection.execute("drop table if exists prototype_identifier_lines")
        connection.execute(
            "create table prototype_identifier_lines ("
            "file_id integer not null, line integer not null, name text not null, "
            "primary key (file_id, line, name)) without rowid"
        )
        connection.execute(
            "create index prototype_identifier_lines_name "
            "on prototype_identifier_lines(name, file_id, line)"
        )
        source_rows = connection.execute("select file_id, line, text from lines")
        with connection:
            for file_id, line, text in source_rows:
                names = identifiers(text)
                if names:
                    connection.executemany(
                        "insert into prototype_identifier_lines(file_id, line, name) "
                        "values (?, ?, ?)",
                        ((file_id, line, name) for name in names),
                    )
    finally:
        connection.close()
    return time.perf_counter_ns() - started


def update_index_file(db_path: Path, relative_path: str) -> int:
    import sqlite3

    started = time.perf_counter_ns()
    connection = sqlite3.connect(db_path)
    try:
        file_row = connection.execute(
            "select id from files where path = ?", (relative_path,)
        ).fetchone()
        if file_row is None:
            raise RuntimeError(f"updated path is not indexed: {relative_path}")
        file_id = file_row[0]
        with connection:
            connection.execute(
                "delete from prototype_identifier_lines where file_id = ?", (file_id,)
            )
            for line, text in connection.execute(
                "select line, text from lines where file_id = ? order by line", (file_id,)
            ):
                names = identifiers(text)
                if names:
                    connection.executemany(
                        "insert into prototype_identifier_lines(file_id, line, name) "
                        "values (?, ?, ?)",
                        ((file_id, line, name) for name in names),
                    )
    finally:
        connection.close()
    return time.perf_counter_ns() - started


def contains_identifier(line: str, name: str) -> bool:
    for start in range(len(line) - len(name) + 1):
        end = start + len(name)
        if line[start:end] != name:
            continue
        if start > 0 and is_identifier_char(line[start - 1]):
            continue
        if end < len(line) and is_identifier_char(line[end]):
            continue
        return True
    return False


def baseline_scan_query(connection, name: str, limit: int) -> list[dict[str, object]]:
    """Match refs' SQL substring prefilter plus Go-style exact boundary check."""
    definition_lines = {
        (file_id, line)
        for file_id, line in connection.execute(
            "select file_id, line from symbols where name = ?", (name,)
        )
    }
    candidates = []
    for row in connection.execute(
        "select files.id as file_id, files.path as path, lines.line as line, "
        "files.language as language, lines.text as text "
        "from lines join files on files.id = lines.file_id "
        "where instr(lines.text, ?) > 0 order by files.path, lines.line",
        (name,),
    ):
        if (row["file_id"], row["line"]) in definition_lines:
            continue
        if not contains_identifier(row["text"], name):
            continue
        candidates.append(dict(row))
        if len(candidates) >= limit:
            break
    return candidates


def prototype_query(connection, name: str, limit: int) -> list[dict[str, object]]:
    return [
        dict(row)
        for row in connection.execute(
            "select files.id as file_id, files.path as path, lines.line as line, "
            "files.language as language, lines.text as text "
            "from prototype_identifier_lines as identifiers "
            "join lines on lines.file_id = identifiers.file_id "
            "and lines.line = identifiers.line "
            "join files on files.id = identifiers.file_id "
            "where identifiers.name = ? and not exists ("
            "select 1 from symbols where symbols.file_id = identifiers.file_id "
            "and symbols.line = identifiers.line and symbols.name = ?) "
            "order by files.path, lines.line limit ?",
            (name, name, limit),
        )
    ]


def size(path: Path) -> int:
    return path.stat().st_size


def candidate_keys(rows: list[dict[str, object]]) -> list[tuple[str, int]]:
    return [(str(row["path"]), int(row["line"])) for row in rows]


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo", required=True, type=Path, help="checked-out source repository")
    parser.add_argument("--mzci", required=True, type=Path, help="built mzci executable")
    parser.add_argument("--work", type=Path, help="temporary work directory (default: system temp)")
    parser.add_argument("--query", default=DEFAULT_QUERY)
    parser.add_argument("--update-path", default=DEFAULT_UPDATE_PATH)
    parser.add_argument("--repeats", type=int, default=7)
    parser.add_argument("--limit", type=int, default=100)
    args = parser.parse_args()

    repo = args.repo.resolve()
    mzci = args.mzci.resolve()
    update_source = repo / args.update_path
    if not (repo / ".git").exists() or not update_source.is_file():
        parser.error("--repo must be a Git checkout containing --update-path")
    if not os.access(mzci, os.X_OK):
        parser.error("--mzci must name an executable")
    if args.repeats < 1 or args.limit < 1:
        parser.error("--repeats and --limit must be positive")

    temp_context = tempfile.TemporaryDirectory(prefix="mzci-refs-index-") if args.work is None else None
    work = Path(temp_context.name) if temp_context else args.work.resolve()
    work.mkdir(parents=True, exist_ok=True)
    base_db = work / "baseline.sqlite"
    update_repo = work / "update-repo"
    update_db = work / "update.sqlite"

    current_rebuild_ns, _ = timed([str(mzci), "rebuild", "--db", str(base_db), str(repo)])
    base_size = size(base_db)
    overlay_build_ns = materialize_index(base_db)
    overlay_size = size(base_db)

    current_query = [
        str(mzci), "refs", "--db", str(base_db), "--details", "--format", "json",
        "--limit", str(args.limit), args.query,
    ]
    current_query_ns, current_output = median_command(current_query, args.repeats)
    current_rows = json.loads(current_output)["candidates"]
    import sqlite3

    connection = sqlite3.connect(base_db)
    connection.row_factory = sqlite3.Row
    try:
        baseline_rows = baseline_scan_query(connection, args.query, args.limit)
        prototype_rows = prototype_query(connection, args.query, args.limit)
        equivalent = (
            candidate_keys(current_rows) == candidate_keys(baseline_rows)
            and candidate_keys(baseline_rows) == candidate_keys(prototype_rows)
        )

        baseline_scan_ns = []
        prototype_query_ns = []
        # Warm both paths once, then alternate in the same process and connection.
        for _ in range(args.repeats):
            started = time.perf_counter_ns()
            baseline_scan_query(connection, args.query, args.limit)
            baseline_scan_ns.append(time.perf_counter_ns() - started)
            started = time.perf_counter_ns()
            prototype_query(connection, args.query, args.limit)
            prototype_query_ns.append(time.perf_counter_ns() - started)
    finally:
        connection.close()

    run(["git", "clone", "--quiet", "--shared", "--no-hardlinks", str(repo), str(update_repo)])
    head = run(["git", "-C", str(repo), "rev-parse", "HEAD"], capture=True).strip()
    run(["git", "-C", str(update_repo), "checkout", "--quiet", "--detach", head])
    update_source_copy = update_repo / args.update_path
    update_db_rebuild_ns, _ = timed(
        [str(mzci), "rebuild", "--db", str(update_db), str(update_repo)]
    )
    update_overlay_build_ns = materialize_index(update_db)

    with update_source_copy.open("a", encoding="utf-8") as source:
        source.write("\n# refs-index benchmark update\n")
    run(["git", "-C", str(update_repo), "add", args.update_path])
    current_update_ns, _ = timed(
        [str(mzci), "update", "--db", str(update_db), "--format", "json", str(update_repo)]
    )
    overlay_update_ns = update_index_file(update_db, args.update_path)

    report = {
        "repo": str(repo),
        "revision": head,
        "query": args.query,
        "update_path": args.update_path,
        "candidates": len(current_rows),
        "query_results_equivalent": equivalent,
        "repeats": args.repeats,
        "baseline": {
            "rebuild_ms": round(current_rebuild_ns / 1_000_000, 2),
            "db_bytes": base_size,
            "refs_cli_median_ms": round(current_query_ns / 1_000_000, 2),
            "refs_in_process_scan_median_ms": round(statistics.median(baseline_scan_ns) / 1_000_000, 2),
        },
        "prototype": {
            "identifier_index_build_ms_additional": round(overlay_build_ns / 1_000_000, 2),
            "projected_rebuild_ms": round((current_rebuild_ns + overlay_build_ns) / 1_000_000, 2),
            "db_bytes": overlay_size,
            "db_size_delta_bytes": overlay_size - base_size,
            "db_size_delta_percent": round((overlay_size - base_size) * 100 / base_size, 2),
            "refs_in_process_median_ms": round(statistics.median(prototype_query_ns) / 1_000_000, 2),
            "identifier_index_update_ms_additional": round(overlay_update_ns / 1_000_000, 2),
            "projected_single_file_update_ms": round((current_update_ns + overlay_update_ns) / 1_000_000, 2),
        },
        "single_file_update": {
            "baseline_update_ms": round(current_update_ns / 1_000_000, 2),
            "prototype_extra_index_update_ms": round(overlay_update_ns / 1_000_000, 2),
            "prototype_db_bytes_after_update": size(update_db),
            "initial_rebuild_ms_for_update_copy": round(update_db_rebuild_ns / 1_000_000, 2),
            "initial_extra_index_build_ms_for_update_copy": round(update_overlay_build_ns / 1_000_000, 2),
        },
        "notes": [
            "The prototype records all identifier-like runs in every source line, including comments and strings.",
            "The fixed query is case-sensitive, matching refs default behavior.",
            "The identifier table is an experimental overlay, not part of mzci's production schema.",
            "Projected times add the measured overlay construction/update time to mzci's measured operation.",
            "The in-process baseline scan and prototype use the same Python process and SQLite connection; the baseline preserves refs' SQL substring prefilter and exact identifier-boundary check.",
            "The end-to-end mzci refs CLI time is reported separately because it includes subprocess and JSON overhead absent from either in-process path.",
        ],
    }
    print(json.dumps(report, indent=2, sort_keys=True))
    if temp_context:
        temp_context.cleanup()
    return 0 if equivalent else 2


if __name__ == "__main__":
    raise SystemExit(main())
