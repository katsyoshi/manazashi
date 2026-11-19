# Fixture tasks for bench/context/test.sh.

TASKS=(def_save long_read broken missing_read empty_json)

def_save_target=lib/a.rb
def_save_with() {
  step defs mz defs save
  step show mz show --line 3 --context 5 lib/a.rb
}
def_save_without() {
  step rg rg -n --sort path "def save"
  step read read_file lib/a.rb
}

long_read_target=lib/long.rb
long_read_with() {
  step outline mz outline lib/long.rb
}
long_read_without() {
  step read read_file lib/long.rb
}

broken_target=lib/a.rb
broken_with() {
  step miss false
}
broken_without() {
  step miss true
}

missing_read_target=lib/missing.rb
missing_read_with() {
  step read read_file lib/missing.rb
}
missing_read_without() {
  step read read_file lib/missing.rb
}

empty_json_target=lib/missing.rb
empty_json_with() {
  step outline mz outline lib/missing.rb
}
empty_json_without() {
  step show mz show --line 1 lib/missing.rb
}
