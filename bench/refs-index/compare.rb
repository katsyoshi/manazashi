#!/usr/bin/env ruby
# frozen_string_literal: true

require "csv"
require "fileutils"
require "json"
require "open3"
require "optparse"
require "pathname"
require "tmpdir"

DEFAULT_QUERY = "assert_valid_keys"
DEFAULT_UPDATE_PATH = "activesupport/lib/active_support/core_ext/hash/keys.rb"
INDEX_TABLE = "prototype_identifier_lines"

def run(*command, capture: false)
  stdout, stderr, status = Open3.capture3(*command)
  raise "#{command.join(' ')} failed (#{status.exitstatus}): #{stderr}" unless status.success?

  capture ? stdout : ""
end

def timed
  started = Process.clock_gettime(Process::CLOCK_MONOTONIC)
  output = yield
  [(Process.clock_gettime(Process::CLOCK_MONOTONIC) - started) * 1_000_000_000, output]
end

def median(values)
  sorted = values.sort
  middle = sorted.length / 2
  return sorted[middle] if sorted.length.odd?

  (sorted[middle - 1] + sorted[middle]) / 2.0
end

def sql_quote(value)
  "'#{value.to_s.gsub("'", "''")}'"
end

def identifier_runs(text)
  text.scan(/[\p{L}\p{Nd}_$]+/u).uniq
end

def each_sqlite_csv_row(db_path, sql)
  Open3.popen3("sqlite3", "-csv", "-noheader", db_path.to_s, sql) do |stdin, stdout, stderr, waiter|
    stdin.close
    stderr_reader = Thread.new { stderr.read }
    CSV.new(stdout).each { |row| yield row }
    status = waiter.value
    error = stderr_reader.value
    raise "sqlite3 CSV query failed (#{status.exitstatus}): #{error}" unless status.success?
  end
end

def with_sqlite_writer(db_path)
  Open3.popen3("sqlite3", db_path.to_s) do |stdin, stdout, stderr, waiter|
    stdout_reader = Thread.new { stdout.read }
    stderr_reader = Thread.new { stderr.read }
    stdin.write("begin;\n")
    yield stdin
    stdin.write("commit;\n")
    stdin.close
    status = waiter.value
    output = stdout_reader.value
    error = stderr_reader.value
    raise "sqlite3 write failed (#{status.exitstatus}): #{error}#{output}" unless status.success?
  end
end

def materialize_index(db_path)
  elapsed, = timed do
    run("sqlite3", db_path.to_s, <<~SQL)
      drop table if exists #{INDEX_TABLE};
      create table #{INDEX_TABLE} (
        file_id integer not null,
        line integer not null,
        name text not null,
        primary key (file_id, line, name)
      ) without rowid;
    SQL
    with_sqlite_writer(db_path) do |writer|
      tuples = []
      each_sqlite_csv_row(db_path, "select file_id, line, text from lines order by file_id, line;") do |file_id, line, text|
        identifier_runs(text).each do |name|
          tuples << "(#{Integer(file_id)}, #{Integer(line)}, #{sql_quote(name)})"
          if tuples.length >= 500
            writer.write("insert into #{INDEX_TABLE}(file_id, line, name) values #{tuples.join(',')};\n")
            tuples.clear
          end
        end
      end
      writer.write("insert into #{INDEX_TABLE}(file_id, line, name) values #{tuples.join(',')};\n") unless tuples.empty?
    end
    run("sqlite3", db_path.to_s,
        "create index #{INDEX_TABLE}_name on #{INDEX_TABLE}(name, file_id, line);")
  end
  elapsed
end

def refresh_index_file(db_path, relative_path)
  elapsed, = timed do
    file_id = run("sqlite3", "-noheader", db_path.to_s,
                  "select id from files where path = #{sql_quote(relative_path)};",
                  capture: true).strip
    raise "updated path is not indexed: #{relative_path}" if file_id.empty?

    with_sqlite_writer(db_path) do |writer|
      writer.write("delete from #{INDEX_TABLE} where file_id = #{Integer(file_id)};\n")
      tuples = []
      each_sqlite_csv_row(db_path,
                          "select line, text from lines where file_id = #{Integer(file_id)} order by line;") do |line, text|
        identifier_runs(text).each do |name|
          tuples << "(#{Integer(file_id)}, #{Integer(line)}, #{sql_quote(name)})"
          if tuples.length >= 500
            writer.write("insert into #{INDEX_TABLE}(file_id, line, name) values #{tuples.join(',')};\n")
            tuples.clear
          end
        end
      end
      writer.write("insert into #{INDEX_TABLE}(file_id, line, name) values #{tuples.join(',')};\n") unless tuples.empty?
    end
  end
  elapsed
end

def identifier_boundary?(char)
  char.match?(/[\p{L}\p{Nd}_$]/u)
end

def contains_identifier?(line, name)
  line_chars = line.each_char.to_a
  name_chars = name.each_char.to_a
  return false if name_chars.empty? || name_chars.length > line_chars.length

  (0..(line_chars.length - name_chars.length)).each do |start|
    next unless line_chars.slice(start, name_chars.length) == name_chars
    next if start.positive? && identifier_boundary?(line_chars[start - 1])

    finish = start + name_chars.length
    next if finish < line_chars.length && identifier_boundary?(line_chars[finish])

    return true
  end
  false
end

def csv_rows(db_path, sql)
  output = run("sqlite3", "-csv", "-noheader", db_path.to_s, sql, capture: true)
  CSV.parse(output)
end

def baseline_scan_query(db_path, name, limit)
  quoted_name = sql_quote(name)
  sql = <<~SQL
    select files.id, files.path, lines.line, files.language, lines.text,
           exists(select 1 from symbols where symbols.file_id = files.id
                  and symbols.line = lines.line and symbols.name = #{quoted_name})
    from lines join files on files.id = lines.file_id
    where instr(lines.text, #{quoted_name}) > 0
    order by files.path, lines.line;
  SQL
  rows = []
  csv_rows(db_path, sql).each do |file_id, path, line, language, text, definition|
    next if definition == "1"
    next unless contains_identifier?(text, name)

    rows << { "file_id" => file_id.to_i, "path" => path, "line" => line.to_i,
              "language" => language, "text" => text }
    break if rows.length >= limit
  end
  rows
end

def prototype_query(db_path, name, limit)
  quoted_name = sql_quote(name)
  sql = <<~SQL
    select files.id, files.path, lines.line, files.language, lines.text
    from #{INDEX_TABLE} as identifiers
    join lines on lines.file_id = identifiers.file_id and lines.line = identifiers.line
    join files on files.id = identifiers.file_id
    where identifiers.name = #{quoted_name} and not exists (
      select 1 from symbols where symbols.file_id = identifiers.file_id
        and symbols.line = identifiers.line and symbols.name = #{quoted_name}
    )
    order by files.path, lines.line limit #{Integer(limit)};
  SQL
  csv_rows(db_path, sql).map do |file_id, path, line, language, text|
    { "file_id" => file_id.to_i, "path" => path, "line" => line.to_i,
      "language" => language, "text" => text }
  end
end

def candidate_keys(rows)
  rows.map { |row| [row.fetch("path"), row.fetch("line").to_i] }
end

def db_size(path)
  File.size(path)
end

def main
  options = {
    query: DEFAULT_QUERY,
    update_path: DEFAULT_UPDATE_PATH,
    repeats: 7,
    limit: 100
  }
  OptionParser.new do |parser|
    parser.banner = "Usage: compare.rb --repo DIR --mzci FILE [options]"
    parser.on("--repo DIR", "checked-out source repository") { |value| options[:repo] = Pathname.new(value).expand_path }
    parser.on("--mzci FILE", "built mzci executable") { |value| options[:mzci] = Pathname.new(value).expand_path }
    parser.on("--work DIR", "retain outputs under this directory") { |value| options[:work] = Pathname.new(value).expand_path }
    parser.on("--query NAME", "identifier to query (default: #{DEFAULT_QUERY})") { |value| options[:query] = value }
    parser.on("--update-path PATH", "tracked file to modify in clone") { |value| options[:update_path] = value }
    parser.on("--repeats N", Integer, "query timing samples (default: 7)") { |value| options[:repeats] = value }
    parser.on("--limit N", Integer, "candidate limit (default: 100)") { |value| options[:limit] = value }
  end.parse!

  abort "--repo and --mzci are required" unless options[:repo] && options[:mzci]
  repo = options[:repo]
  mzci = options[:mzci]
  update_source = repo.join(options[:update_path])
  abort "--repo must be a Git checkout containing --update-path" unless repo.join(".git").exist? && update_source.file?
  abort "--mzci must name an executable" unless mzci.file? && File.executable?(mzci)
  abort "--repeats and --limit must be positive" unless options[:repeats].positive? && options[:limit].positive?

  temporary_work = options[:work].nil?
  work = options[:work] || Pathname.new(Dir.mktmpdir("mzci-refs-index-"))
  at_exit { FileUtils.remove_entry(work) if temporary_work && work.exist? }
  run_dir = work.join("run-#{Time.now.utc.strftime('%Y%m%dT%H%M%S')}-#{Process.pid}")
  run_dir.mkpath
  base_db = run_dir.join("baseline.sqlite")
  update_repo = run_dir.join("update-repo")
  update_db = run_dir.join("update.sqlite")

  rebuild_ns, = timed { run(mzci.to_s, "rebuild", "--db", base_db.to_s, repo.to_s) }
  base_size = db_size(base_db)
  overlay_build_ns = materialize_index(base_db)
  overlay_size = db_size(base_db)

  current_command = [mzci.to_s, "refs", "--db", base_db.to_s, "--details", "--format", "json",
                     "--limit", options[:limit].to_s, options[:query]]
  current_query_ns = []
  current_output = nil
  options[:repeats].times do
    duration, output = timed { run(*current_command, capture: true) }
    current_query_ns << duration
    current_output = output
  end
  current_rows = JSON.parse(current_output).fetch("candidates")

  baseline_rows = baseline_scan_query(base_db, options[:query], options[:limit])
  prototype_rows = prototype_query(base_db, options[:query], options[:limit])
  equivalent = candidate_keys(current_rows) == candidate_keys(baseline_rows) &&
               candidate_keys(baseline_rows) == candidate_keys(prototype_rows)
  baseline_scan_ns = []
  prototype_query_ns = []
  options[:repeats].times do
    duration, = timed { baseline_scan_query(base_db, options[:query], options[:limit]) }
    baseline_scan_ns << duration
    duration, = timed { prototype_query(base_db, options[:query], options[:limit]) }
    prototype_query_ns << duration
  end

  run("git", "clone", "--quiet", "--shared", "--no-hardlinks", repo.to_s, update_repo.to_s)
  revision = run("git", "-C", repo.to_s, "rev-parse", "HEAD", capture: true).strip
  run("git", "-C", update_repo.to_s, "checkout", "--quiet", "--detach", revision)
  update_db_rebuild_ns, = timed { run(mzci.to_s, "rebuild", "--db", update_db.to_s, update_repo.to_s) }
  update_overlay_build_ns = materialize_index(update_db)

  File.open(update_repo.join(options[:update_path]), "a", encoding: "UTF-8") do |source|
    source.write("\n# refs-index benchmark update\n")
  end
  run("git", "-C", update_repo.to_s, "add", options[:update_path])
  update_ns, = timed do
    run(mzci.to_s, "update", "--db", update_db.to_s, "--format", "json", update_repo.to_s)
  end
  overlay_update_ns = refresh_index_file(update_db, options[:update_path])

  report = {
    repo: repo.to_s,
    revision: revision,
    query: options[:query],
    update_path: options[:update_path],
    candidates: current_rows.length,
    query_results_equivalent: equivalent,
    repeats: options[:repeats],
    baseline: {
      rebuild_ms: (rebuild_ns / 1_000_000.0).round(2),
      db_bytes: base_size,
      refs_cli_median_ms: (median(current_query_ns) / 1_000_000.0).round(2),
      refs_csv_scan_median_ms: (median(baseline_scan_ns) / 1_000_000.0).round(2)
    },
    prototype: {
      identifier_index_build_ms_additional: (overlay_build_ns / 1_000_000.0).round(2),
      projected_rebuild_ms: ((rebuild_ns + overlay_build_ns) / 1_000_000.0).round(2),
      db_bytes: overlay_size,
      db_size_delta_bytes: overlay_size - base_size,
      db_size_delta_percent: ((overlay_size - base_size) * 100.0 / base_size).round(2),
      refs_csv_median_ms: (median(prototype_query_ns) / 1_000_000.0).round(2),
      identifier_index_update_ms_additional: (overlay_update_ns / 1_000_000.0).round(2),
      projected_single_file_update_ms: ((update_ns + overlay_update_ns) / 1_000_000.0).round(2)
    },
    single_file_update: {
      baseline_update_ms: (update_ns / 1_000_000.0).round(2),
      prototype_extra_index_update_ms: (overlay_update_ns / 1_000_000.0).round(2),
      prototype_db_bytes_after_update: db_size(update_db),
      initial_rebuild_ms_for_update_copy: (update_db_rebuild_ns / 1_000_000.0).round(2),
      initial_extra_index_build_ms_for_update_copy: (update_overlay_build_ns / 1_000_000.0).round(2)
    },
    notes: [
      "The prototype records all identifier-like runs in every source line, including comments and strings.",
      "The fixed query is case-sensitive, matching refs default behavior.",
      "The identifier table is an experimental overlay, not part of mzci's production schema.",
      "Projected times add measured overlay construction/update time to mzci's measured operation.",
      "The baseline CSV scan and prototype query use the same sqlite3 CSV subprocess and Ruby CSV parser; the baseline then applies exact identifier boundaries in Ruby.",
      "The end-to-end mzci refs CLI time is reported separately because it includes JSON output and its own subprocess path.",
      "The database size is the main SQLite file, excluding transient sidecars."
    ]
  }
  puts JSON.pretty_generate(report)
  equivalent ? 0 : 2
end

exit(main)
