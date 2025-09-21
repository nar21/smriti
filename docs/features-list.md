# Smriti Database Archival Tool Features


## Features
### Fully customizable archival plan
The archival plan is a Yaml file in which parameters of the archival can be specified. An archival plan is required for each table. 

Check [Archival Plan](archival-plan.md) docs.

### Data consistency [To be implemented]
TODO: Smriti should ensure that the rows which are exported are not modified by other transactions meanwhile.
### Dry run your archival plan
Dry Run prints the plan that Smriti will use to perform the archival activity and exits. The plan contains useful information like execution ID, data file paths, number of chunks that can be inspected before beginning the execution.

A dry run generates an execution ID and a corresponding state file. These can be used to run the actual execution using the `--execution-id` CLI argument when running Smriti.

### Resume your previous execution
Smriti has been designed to handle failures gracefully. Each archival task comprises of multiple stages. In case a stage fails, it is marked as "failed" in the state file.
When Smriti resumes the execution next time, it attempts to execute the failed jobs, and skips previously successful jobs.

### Data isolation within working directory
### Parallel execution
### Compression of data files
### [Optional] Upload to object stores
### [Optional] Auto cleanup of intermediate files
### Dual channel log output [To be implemented]

## Supported database engines
1. Postgresql

## Supported object stores
1. S3