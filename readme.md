# How to use this tool

## preparation

+ hosts.txt

this file contains the ip addresses of the nodes that running the tss benchmark.
current we support digital ocean machine (username root) and AWS machines.
we put an `id` as `d` as the digital ocean machine and `a` as the AWS machine.
the fist node in the hosts.txt is taken as the bootstrap node.



+ local configuration folders

we need to creat folders as the fime name `1,2,3,4` under `storage` folder.


## How to Run
the binary is under cmd folder.

**num** indicates how many nodes you want to create the configuration file.

**init** indicates whether you want to recreate all the configuration file. we usually set it true when we have new
nodes and want to create the configuration for new nodes.


`opt` 

1 for upload the configuration files to cloud machines.

2 for run keygen test

3 for run keysign test


