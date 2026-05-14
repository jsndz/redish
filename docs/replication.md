Replication:

Replication is a creating other redis 
with their own store clients and config

they are connected to the main through a tcp connection

Replicas connect to master 
Replica purpose:

read scaling
backup redundancy
failover
data durability

after the port config and replcationof config in the cli

we need to do handshake between master and slave

the slave sends the following for the handshake:

A PING from the replica
REPLCONF twice from the replica
PSYNC from the replica

REPLCONF listening-port <PORT>: This tells the master which port the replica is listening on. This value is used for monitoring and logging, not for replication itself.
REPLCONF capa psync2: This notifies the master of the replica's capabilities.
capa stands for "capabilities". It indicates that the next argument is a feature the replica supports.
psync2 signals that the replica supports the PSYNC2 protocol. PSYNC2 is an improved version of the partial synchronization feature used to resynchronize a replica with its master.


PSYNC <replId> <offset> will be sent by slave
for the first time slave wont know replId so it will send ? -1
Server responds with FULLRSYNC replId offset

FULLRESYNC means that the master cannot perform an incremental update to the replica, and will start a full resynchronization.

so for every write command sent to the master, it will be sent to replica's 
replica's wont send any response back if the command is from master which has a separate handler

this might be seen as a asynchrous and less durable and less persistance
since the replica may or may not be updates since the master does not wait for reply

wait command is used to make this whole thing more durable
after the command is writtern and use WAIT numOfReplicas timeout
then the master waits for ack for min num of replicas for the timeout