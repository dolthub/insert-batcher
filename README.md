# Insert Batcher

This is a simple tool that takes a SQL dump file, and combines inserts
provided a batching factor.

Example input:
```sql
create table xy (x int primary key,y int);
INSERT INTO xy values (0,'0');
INSERT INTO xy values (1,'1');
INSERT INTO xy values (2,'2');
INSERT INTO xy (y,x) values ('4',4);
INSERT INTO xy (y,x) values ('5',5);
INSERT INTO xy (y,x) values ('6',6);
```

example output:
```sql
create table xy (
	x int primary key,
	y int
);
insert into xy values (0, '0'), (1, '1'), (2, '2');
insert into xy(y, x) values ('4', 4), ('5', 5), ('6', 6);
```

The algorithm is very simple, does not consider foreign key ordering,
and will directly passthrough rows that do not match the narrow
`INSERT INTO ... VALUES ...` pattern. For example, an insert like
`INSERT INTO ... SELECT ...` in the middle of a series of `VALUES`
inserts will be written as encountered, potentially reordering the
inserts:
```sql
-- before
create table xy (x int primary key, y int);
INSERT INTO xy values (0,'0');
INSERT INTO xy values (1,'1');
INSERT INTO xy values (2,'2');
INSERT INTO xy (y,x) values ('4',4);
INSERT INTO xy (y,x) values ('5',5);
INSERT INTO xy select x+1, y+1 from xy where x > 0;

-- after
insert into xy select x + 1, y + 1 from xy where x > 0;
insert into xy values (0, '0'), (1, '1'), (2, '2');
insert into xy(y, x) values ('4', 4), ('5', 5);
```

Clearly, the "before" and "after" are not logically equivalent in
this case.