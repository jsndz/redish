The commands that have been added:

1. echo HELLO
-> HELLO

2. set key val
sets up a map with key and val
everything is stored as string 
and internal conversion happens
-> HELLO

3. get key
to get the value for key
-> val

4. incr key
increment the value of the key should be a number
if key is not there set to 1 
-> 1(incremented val)

5. ping
-> pong

6. rpush mylist val1 val2 etc
push to list to right
creates a list and appends to it
-> 2 (len of list)

7. lrange key 0 5
(list range)
gives the element is the list form start to end like 0 to 5
-> 1)a

