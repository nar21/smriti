# Smriti Database Archival Tool Features


## Features
### Fully customizable archival plan
The archival plan is a Yaml file in which parameters of the archival can be specified by the user. Each table in your database requires its own archival plan file. This file does not contain any secrets, therefore it can be checked in to Git.

Check [Archival Plan](archival-plan.md) docs.

### Data consistency [To be implemented]
TODO: Smriti should ensure that the rows which are exported are not modified by other transactions meanwhile.

### Dry run your archival plan
When you run Smriti in dry run mode, it prints the plan will be used to perform the archival activity and exits. The plan contains useful information like execution ID, data file paths and the number of data chunks. The plan can be inspected before starting the execution.

A dry run generates an execution ID and a corresponding state file. These can be used to run the actual execution using the `--execution-id` CLI argument when running Smriti.

### Resume your previous execution
Smriti has been designed to handle failures gracefully. If batching is enabled, Smriti creates multiple jobs automatically. A job processes the data range that it is assigned. 

Each job comprises of multiple stages. In case a stage fails, it is marked as "failed" in the state file. When you run Smriti with the execution ID next time, it attempts to execute the failed jobs from the latest failed stage. On the other hand, it skips previously successful stages and jobs.

### Data isolation within working directory
Every new archival execution creates an execution ID, and a corresponding directory within the `workDir`. A resumed archival uses files in an already existing directory. This data isolation makes troubleshooting, log collection easier and also simplifies the code.

### Parallel execution [To be implemented]
An archival plan is transformed into one or more jobs, created and managed internally by Smriti. The number of jobs and number of workers can be specified in the archival plan. A job is mapped to a worker thread which then processes a subset of the data.

### Compression of data files
The exported data is stored in flat plaintext files like CSV. When archiving large tables, CSV file sizes can run into multiple GBs. This demands more disk and network resources and also increases storage costs.

Smriti's compression stage compresses the data files using popular compression algorithms to achieve resource and cost efficiency.

### [Optional] Upload to object stores
Smriti supports uploading of exported and compressed data file to object storage services like S3. It is possible to set vendor specific options in the archival plan.

### [Optional] Auto cleanup of intermediate files
As an archival solution, Smriti works with large volumes of data. One of Smriti's aims is to keep resource utilization low while performing archival. One way it achieves this is by optionally cleaning up intermediate data files after they are successfully uploaded to object storage.

### Dual channel log output [To be implemented]
Output of the main thread and each worker thread is printed to both the screen  `stdout` and their respective log files within the execution directory.

## Supported database engines
1. Postgresql

## Supported object stores
1. S3