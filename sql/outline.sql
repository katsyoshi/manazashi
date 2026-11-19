with target as (
  select id, path
  from files
  where path = %s or path like %s
  order by case when path = %s then 0 else 1 end, length(path)
  limit 1
)
select %s
from symbols join target on target.id = symbols.file_id
order by symbols.line, symbols.column, symbols.name;
