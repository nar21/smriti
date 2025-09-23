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
make
```

## Usage

### Create your archival plan
1. Create a copy of the sample archival plan 
``` shell
cp archival-plan/sample.yaml archival-plan/myplan.yaml
```

Refer to the [Archival Plan documentation](docs/archival-plan.md) to understand the parameters.

Your archival plans can be checked into SCM, as credentials are stored in a separate file `./db-credentials.yaml` .

2. Initialize Smriti
`--init` creates settings.yaml and db-credentials.yaml files with default values
``` shell
./bin/smriti --init
```

Notes: 
  - All database credentials should be set in the `db_credentials.yaml` file.
  - Update the default paths in settings.yaml as per your convenience.



3. Set ARCHIVAL_PLAN environment variable

This is the name of the archival plan file. Specify only the filename; file extension and path must not be included.
``` shell
export ARCHIVAL_PLAN="sample"
```

### Execute your archival
1. Dry run your archival plan
``` shell
./bin/smriti --dry-run
```


2. Start archival
- Fresh execution
``` shell
./bin/smriti 
```
- Resume with execution ID
``` shell
./bin/smriti --execution-id <exec-id>

```


If all goes well, you should see archival files created in the working directory `workDir/` .
