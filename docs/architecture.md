# Smriti Database Archival Tool Architecure


### Concepts
1. Execution ID: After Smriti begins a new execution, it generates a unique execution ID. This ID isolates the data files, logs and any other metadata of the current archival activity from others. 

    The execution ID can be used to resume a previous execution if it failed for some reason.

2. Working Directory: A working directory `./workDir` is created if not present in the current directory. Directories for each execution, identified by its ID, will be created within the working directory.

3. Statefile: The statefile stores the current execution state of all worker threads. This statefile is basically a dump of the Archival Plan struct. This file is used to resume an older execution. It is stored within the execution directory.

4. Dry Run: The code outlines the plan of the archival that it will execute. An execution which has just dry run can be resumed, as an execution ID is generated even on dry run.

5. Workers (To be implemented): The number of parallel workers to be spawned. Each worker will archive its data range and will open an individual connection to the database. Workers are fault isolated.