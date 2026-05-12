Let's Say there are 2 clients A and B

A is doing transaction :

MULTI 
SET balance 10
GET code
SET balance code +10
EXEC

and B changes the balance to 20 after the GET code command

Then the value will be invalid since A is changing based on balance 10
the whole thing is big ERROR

To solve this we use Optimistic locking using command WATCH

SO in simple terms watch 

Before starting transaction you WATCH the variables you want to change
And if they change your transaction wont run

This is optimistic because you allow others to edit the value of the key 
and you yourself are stopping the transaction

It would be pessimistic locking if you wouldn't let others edit it

this process is done by having 
watched keys in store (key -> clients)
watched keys in client (client -> keys)

both use map because we want fast lookup and delete

When we call watch we add watchers to the key by adding both in store and client
and when the key is deleted remove watchers
and call touch watchers when the set function is called 
if the key has any watchers then it is set as dirty
 and transaction wont run if it is dirty

 there are also commands like unwatch discard 