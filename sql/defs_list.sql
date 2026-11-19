select %s
from symbols
where %s
order by path, line, column, name
limit %d;
