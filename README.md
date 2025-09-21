## Smriti Database Archival Tool

Smriti is a tool to selectively archive data from live databases to plaintext file formats like CSV, and optionally upload them to object stores like S3.

⚠️ Smriti is currently at `alpha` stability level.

## Full docs index

1. [Architecture](docs/architecture.md)
2. [Smriti Features List](docs/features-list.md)

## Installation

### Compile from source (requires Golang 1.23)

``` shell
git clone git@github.com:nar21/smriti.git
```

``` shell
cd smriti
go build -o bin/smriti .
```

## Usage

### Create your archival plan
1. Create a copy of the sample archival plan 
``` shell
cp archival-plan/sample.yaml archival-plan/myplan.yaml
```

Refer to the [Archival Plan documentation](docs/archival-plan.md) to understand the parameters.

Your archival plans can be checked into SCM, as credentials are stored in a separate file `./db_credentials.yaml` .

2. Once the archival plan is ready, set the database credentials.
First create the credentials file
``` shell
touch db_credentials.yaml
```

Add the below Yaml block in the `db_credentials.yaml` file and update the appropriate values. 
``` yaml
databases:
  sample-rds:
    engine: "postgres"
    host: localhost
    port: 54321
    user: sample
    password: sample
    dbname: sample
```

Note: All database credentials should be set in the `db_credentials.yaml` file. This file will not be tracked by SCM.

3. Set ARCHIVAL_PLAN environment variable

``` shell
export ARCHIVAL_PLAN="sample"
```

### Execute your archival
1. (Preferably always) Dry run your archival
``` shell
./bin/smriti --dry-run
```


2. Start archival 
- Fresh execution
``` shell
./bin/smriti 
```
- with execution ID
``` shell
./bin/smriti --execution-id <exec-id>

```


If all goes well, you should see archival files created in the working directory `workDir/` .