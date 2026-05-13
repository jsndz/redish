Append only file

if redis crashes all data will be lost so you use AOF file to store it.

so there are config you set when you start the redis 
$ redis --dir <dir> --appendonly yes --appenddirname <append_dir_name> --appendfilename <append_file_name>

if appendonly is true then commands will be appended 
and then when starting the server you can pass config with file path to get all the commands run through the append file

manifest file holds all the series of aof file path so that you can add them incrementally