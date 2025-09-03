import pg8000
import random
import string
import uuid
import ipaddress
import datetime
import json
import os



# --- Helpers ---
def random_string(n=10):
    return ''.join(random.choices(string.ascii_letters + string.digits, k=n))

def random_date(start_year=2000, end_year=2030):
    start = datetime.date(start_year, 1, 1)
    end = datetime.date(end_year, 12, 31)
    delta = end - start
    return start + datetime.timedelta(days=random.randint(0, delta.days))

def random_timestamp():
    start = datetime.datetime(2000, 1, 1)
    end = datetime.datetime(2030, 12, 31)
    delta = end - start
    return start + datetime.timedelta(seconds=random.randint(0, int(delta.total_seconds())))

def random_json():
    return {
        "id": random.randint(1, 1000),
        "name": random_string(5),
        "tags": [random_string(3) for _ in range(3)]
    }

def random_mac():
    return ":".join(["%02x" % random.randint(0, 255) for _ in range(6)])


# --- Enum choices ---
enum_choices = ["small", "medium", "large"]


# --- Row generator (without the autoincrement ID) ---
def generate_row():
    return [
        random.randint(-2147483648, 2147483647),  # int
        random.randint(-9223372036854775808, 9223372036854775807),  # bigint
        round(random.uniform(-1e6, 1e6), 4),  # numeric
        random.uniform(-1e3, 1e3),  # real
        random.uniform(-1e6, 1e6),  # double
        random_string(20),  # text
        random_string(15),  # varchar
        random_string(1),   # char
        random.choice([True, False]),  # boolean
        random_date(),  # date
        random_timestamp(),  # timestamp
        random_timestamp(),  # timestamptz
        datetime.time(random.randint(0, 23), random.randint(0, 59)),  # time
        datetime.timedelta(days=random.randint(1, 100)),  # interval
        uuid.uuid4(),  # uuid
        json.dumps(random_json()),  # json
        json.dumps(random_json()),  # jsonb
        "{" + ",".join([str(random.randint(1,100)) for _ in range(4)]) + "}",  # int[]
        "{" + ",".join(['"' + random_string(3) + '"' for _ in range(5)]) + "}",  # text[]
        str(ipaddress.IPv4Address(random.randint(0, 2**32 - 1))),  # inet
        random_mac(),  # macaddr
        #os.urandom(8),  # bytea (raw bytes)
        '\xcf\x86\x8d\x8aWTc\x86',
        # random.choice(enum_choices)  # enum
    ]


# --- Main insert logic ---
def insert_rows(cursor, n=10):

    INSERT_BATCH_SIZE = 1000

    insert_sql = """INSERT INTO sample_data (
        int_col, bigint_col, numeric_col, real_col, double_col,
        text_col, varchar_col, char_col,
        bool_col, date_col, timestamp_col, timestamptz_col,
        time_col, interval_col, uuid_col,
        json_col, jsonb_col,
        int_array_col, text_array_col,
        inet_col, macaddr_col, bytea_col
    )
    VALUES """

    values_sql = """(
            %s, %s, %s, %s, %s,
            '%s', '%s', '%s',
            %s, '%s', '%s', '%s',
            '%s', '%s', '%s',
            '%s', '%s',
            '%s', '%s',
            '%s', '%s', '%s'
        )"""
    
    values_list = []
    batch_ctr = 1
    for i in range(n):
        row = generate_row()
        values_sql_str = values_sql % tuple(row)
        values_list += [values_sql_str]

        # Execute in batches to avoid too large queries
        if len(values_list) >= INSERT_BATCH_SIZE:
            query = insert_sql + " " + ",\n".join(values_list) + ";"
            print("Executing batch: %d" % batch_ctr)
            batch_ctr += 1
            cursor.execute(query)
            conn.commit()
            values_list = []

    
    print(f"Inserted {n} rows successfully.")


if __name__ == "__main__":

    ROWS_TO_INSERT = 1000000
    conn = pg8000.connect(
        user="sample",
        password="sample",
        host="localhost",
        port=54321,
        database="sample"
    )
    cursor = conn.cursor()
    insert_rows(cursor, ROWS_TO_INSERT)
    cursor.close()
    conn.close()

