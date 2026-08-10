require "json"

exit 2 unless defined?(RubyVM::AbstractSyntaxTree)
major, minor = RUBY_VERSION.split(".", 3).first(2).map(&:to_i)
exit 2 if major < 3 || (major == 3 && minor < 4)

def node_record(node)
  type = node.type.to_s
  record = {
    type: type,
    line: node.first_lineno,
    column: node.first_column,
    end_line: node.last_lineno
  }
  case type
  when "DEFN"
    record[:name] = node.children[0].to_s
  when "DEFS"
    record[:name] = node.children[1].to_s
    record[:receiver] = true
  when "CDECL"
    name = node.children[0]
    return nil unless name.is_a?(Symbol)
    record[:name] = name.to_s
  end
  record
end

ARGF.each_line do |line|
  request = JSON.parse(line)
  nodes = []
  stack = [RubyVM::AbstractSyntaxTree.parse_file(request.fetch("source_path"))]
  until stack.empty?
    node = stack.pop
    if %i[MODULE CLASS DEFN DEFS CDECL].include?(node.type)
      record = node_record(node)
      nodes << record if record
    end
    children = node.children.select { |child| child.is_a?(RubyVM::AbstractSyntaxTree::Node) }
    stack.concat(children.reverse)
  end
  puts JSON.generate(path: request.fetch("path"), ok: true, nodes: nodes)
rescue SyntaxError, StandardError
  puts JSON.generate(path: request&.fetch("path", nil), ok: false, nodes: [])
end
