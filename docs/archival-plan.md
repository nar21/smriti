# Smriti Database Archival Tool 

## Archival Plan File Parameters

1. `databaseID` : This selects the database from which data is to be retrieved. The `databaseID` corresponds to an entry in the `db_credentials.yaml` file.

2. `query`: A dictionary with options to define the data to be retrieved.

    - `table`: The table whose data has to be archived.
    - `filterConditions`: Users may supply a list of SQL expressions representing filter conditions. These expressions are automatically joined with the logical AND operator to form the WHERE clause. This ensures that only rows matching all specified conditions are selected.
    - `batchingEnabled`: Enabled batching which splits the data range into smaller chunks.
    - `batchingColumn`: The column in the table whose values will be used for determining each chunk's start and end values. This column can either by an `int` (or its peers), or a date or date time type.
    - `batchingColumnType`: The data type of the `batchingColumn`. Accepts values either `int` or `date`.
    - `batchStep`: The number of units of the `batchingColumnType` to increase in each subsequent batch's range values.
    - `batchColumnMin`: The minimum value of the batching column.
    - `batchColumnMax`: The maximum value of the batching column.
    
3. `archiveStorage`: Defines the object storage where the files have to be uploaded. Options specific to each vendor should be specified in the respective block.

    - `enabled`: Choose whether to upload the data files to object storage or not.
    - `type`: The service provider of the object storage
    - `s3`: AWS S3 specific options.
        - `bucket`: The S3 bucket to which the files have to be uploaded.
        - `region`: The region in which the S3 bucket resides
    - `local`: A locally mounted filesystem path to which files have to copied. Useful when archives are stored on a separate volume.
        - `path`: The path to which the data files are to be copied.

4. `cleanup`: This section provides flags to modify the behaviour of the Clean up stage.
    - `enabled`: If enabled, deletes the intermediate files after uploading them to object storage. If disabled, retains the intermediate files.