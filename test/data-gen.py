import csv
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
    return json.dumps({
        "id": random.randint(1, 1000),
        "name": random_string(5),
        "tags": [random_string(3) for _ in range(3)]
    })

def random_mac():
    return ":".join(["%02x" % random.randint(0, 255) for _ in range(6)])

# --- Column definitions (simulate Postgres types) ---
columns = [
    "int_col", "bigint_col", "numeric_col", "real_col", "double_col",
    "text_col", "varchar_col", "char_col",
    "bool_col",
    "date_col", "timestamp_col", "timestamptz_col", "time_col", "interval_col",
    "uuid_col",
    "json_col", "jsonb_col",
    "int_array_col", "text_array_col",
    "inet_col", "macaddr_col",
    "bytea_col",
    "enum_col"
]

enum_choices = ["small", "medium", "large"]

# --- Row generator ---
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
        random_date().isoformat(),  # date
        random_timestamp().isoformat(sep=" "),  # timestamp
        random_timestamp().isoformat(sep=" ") + "+00",  # timestamptz
        datetime.time(random.randint(0, 23), random.randint(0, 59)).isoformat(),  # time
        f"{random.randint(1,100)} days",  # interval (string sim)
        str(uuid.uuid4()),  # uuid
        random_json(),  # json
        random_json(),  # jsonb
        "{" + ",".join(str(random.randint(1,100)) for _ in range(5)) + "}",  # int[]
        "{" + ",".join(random_string(3) for _ in range(5)) + "}",  # text[]
        str(ipaddress.IPv4Address(random.randint(0, 2**32 - 1))),  # inet
        random_mac(),  # macaddr
        os.urandom(8).hex(),  # bytea (hex encoding)
        f'"{random.choice(enum_choices)}"'
    ]

# --- Write CSV ---
def generate_csv(filename="pg_types.csv", rows=1):
    with open(filename, "w", newline="") as f:
        writer = csv.writer(f)
        writer.writerow(columns)
        for _ in range(rows):
            writer.writerow(generate_row())

if __name__ == "__main__":
    generate_csv("pg_types.csv", rows=1)
    print("CSV generated: pg_types.csv")
