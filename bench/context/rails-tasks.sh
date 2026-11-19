# Rails context-size tasks for bench/context/run.sh.
# Pinned to rails/rails b6cb9ea7657ea46eb35fdd73b9b2174f7a93a8c0; line numbers
# below are valid only for that revision.

TASKS=(def_body refs find_file outline method_in_class def_body_ambiguous)

# Find ActiveRecord::Persistence#save and read its body. Both sides search
# activerecord/lib; the without side reads the file holding the intended hit.
def_body_target=activerecord/lib/active_record/persistence.rb
def_body_with() {
  step defs mz defs --path activerecord/lib save
  step show mz show --line 415 --context 20 activerecord/lib/active_record/persistence.rb
}
def_body_without() {
  step rg rg -n --sort path 'def save\b' activerecord/lib
  step read read_file activerecord/lib/active_record/persistence.rb
}

# List callers of Hash#assert_valid_keys across the repository.
refs_target=activesupport/lib/active_support/core_ext/hash/keys.rb
refs_with() {
  step refs mz refs assert_valid_keys
}
refs_without() {
  step rg rg -n --sort path -w assert_valid_keys
}

# Locate the routing RouteSet implementation by file name.
find_file_target=actionpack/lib/action_dispatch/routing/route_set.rb
find_file_with() {
  step files mz files route_set
}
find_file_without() {
  step ls-files sh -c 'git ls-files | grep route_set'
}

# Understand what String inflection extensions exist.
outline_target=activesupport/lib/active_support/core_ext/string/inflections.rb
outline_with() {
  step outline mz outline activesupport/lib/active_support/core_ext/string/inflections.rb
}
outline_without() {
  step read read_file activesupport/lib/active_support/core_ext/string/inflections.rb
}

# Read ActionController::Metal#dispatch when the file is already known.
method_in_class_target=actionpack/lib/action_controller/metal.rb
method_in_class_with() {
  step outline mz outline actionpack/lib/action_controller/metal.rb
  step show mz show --line 249 --context 15 actionpack/lib/action_controller/metal.rb
}
method_in_class_without() {
  step read read_file actionpack/lib/action_controller/metal.rb
}

# Find ActiveModel::Errors#initialize among many same-named definitions.
def_body_ambiguous_target=activemodel/lib/active_model/errors.rb
def_body_ambiguous_with() {
  step defs mz defs --path activemodel/lib initialize
  step show mz show --line 115 --context 20 activemodel/lib/active_model/errors.rb
}
def_body_ambiguous_without() {
  step rg rg -n --sort path 'def initialize' activemodel/lib
  step read read_file activemodel/lib/active_model/errors.rb
}
