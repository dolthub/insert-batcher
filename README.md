# Insert Batcher

This is a simple tool that takes a SQL dump file, and combines inserts
provided a batching factor.

## Example

```bash
$ go install github.com/max-hoffman/insert-batcher
$ echo <<EOF
create table xy (x int primary key,y int);
INSERT INTO xy values (0,'0');
INSERT INTO xy values (1,'1');
INSERT INTO xy values (2,'2');
INSERT INTO xy (y,x) values ('4',4);
INSERT INTO xy (y,x) values ('5',5);
INSERT INTO xy (y,x) values ('6',6);
EOF > dump.sql
$ go run github.com/max-hoffman/insert-batcher \
    -in dump.sql \
    -out batched.sql \
    -b 3
$ cat batched.sql
create table xy (
	x int primary key,
	y int
);
insert into xy values (0, '0'), (1, '1'), (2, '2');
insert into xy(y, x) values ('4', 4), ('5', 5), ('6', 6);
```

Example `dump.sql`:
```sql
create table xy (x int primary key,y int);
INSERT INTO xy values (0,'0');
INSERT INTO xy values (1,'1');
INSERT INTO xy values (2,'2');
INSERT INTO xy (y,x) values ('4',4);
INSERT INTO xy (y,x) values ('5',5);
INSERT INTO xy (y,x) values ('6',6);
```

Example `batched.sql`:
```sql
create table xy (
	x int primary key,
	y int
);
insert into xy values (0, '0'), (1, '1'), (2, '2');
insert into xy(y, x) values ('4', 4), ('5', 5), ('6', 6);
```

## Notes

The algorithm is very simple, does not consider foreign key ordering,
and rows that do not match the narrow
`INSERT INTO ... VALUES ...` pattern pass through directly.

For example, an insert like
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

The "before" and "after" databases are not equivalent in this case.