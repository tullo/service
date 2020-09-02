# Cloud SQL

Currently (2020-09-02) it is NOT possible to create an instance using a specific character set and collation with gcloud.

## Workaround 1

1. Google Cloud Platform Console > SQL
2. Create PostgreSQL instance
3. Open instance details > Connect to this instance : `Connect using Cloud Shell`
4. Enter instance credentials
5. Run:

    ```sql
    CREATE DATABASE "example_db"
    WITH OWNER "postgres"
    ENCODING 'UTF8' LC_COLLATE='da_DK.UTF-8' LC_CTYPE='da_DK.UTF-8'
    TEMPLATE template0;
    ```

6. See the database in GCP console> SQL: Databases

> You can only use **`template0`** to create new database with different encoding and locale.

---

## Workaround 2: Drop/Create [Template Databases](https://www.postgresql.org/docs/12/manage-ag-templatedbs.html)

> template1 and template0 do not have any special status beyond the fact that the name `template1` is the *default source database* name for `CREATE DATABASE`.
>
> The `postgres` database is created when a database cluster is initialized. This database is meant as a `default database` for users and applications to connect to.
>
> It is simply a copy of template1 and can be dropped and recreated if necessary.

Step (1) - Create the database instance:

```sh
# (single zone, small, hdd, private address)
gcloud beta sql instances create <DATABASE_INSTANCE> \
  --availability-type=zonal \
  --database-version=POSTGRES_12 \
  --network=default \
  --no-assign-ip \
  --root-password=<PASSWD> \
  --storage-type=HDD \
  --tier=db-f1-micro \
  --zone=<ZONE>
```

Resulting database with default props:

* Character set: `UTF8`
* Collation: `en_US.UTF8`

Step (2) - Replace `template1` using the desired character set and collation:

> Another common reason for copying `template0` instead of template1 is that new `encoding` and `locale` settings can be specified when copying template0, ...
>
> To create a database by copying template0, use:
>
> ```sql
>  CREATE DATABASE dbname TEMPLATE template0;
> ```
>
> https://www.postgresql.org/docs/12/manage-ag-templatedbs.html

```sql
ALTER database template1 is_template=false;

DROP database template1;

CREATE DATABASE template1
    WITH OWNER = postgres
    ENCODING = 'UTF8' LC_COLLATE = 'da_DK.UTF-8' LC_CTYPE = 'da_DK.UTF-8'
    CONNECTION LIMIT = -1
--  TABLESPACE = pg_default (causes an error if enabled)
    TEMPLATE template0;

ALTER database template1 is_template=true;

-- examine or update current state:
-- SELECT * from pg_database;
-- UPDATE pg_database SET datistemplate=true WHERE datname='template1';
-- https://stackoverflow.com/questions/18870775/how-to-change-the-template-database-collection-coding
```

---

## SQL Proxy - Connect to Cloud SQL without allowed networks

> GCE VM -> GCP Firewall -> Cloud SQL Instance **(Working)**
>
> Laptop -> VPN -> On-Prem Firewall -> GCP Firewall -> Cloud SQL Instance **(NOT working)**

The workaround I found is the below:

Steps to be done in **Google Cloud**:

1. Enable [SQl Admin Api](https://console.cloud.google.com/apis/api/sqladmin.googleapis.com/) for the project your instance is part of
2. Give instance a `Public IP`: Edit SQL instance > Connectivity > Public IP > Save
3. Do **NOT** authorize any external networks

Steps to be done locally on your laptop/machine:

1. Install [gcloud](https://cloud.google.com/sdk/docs/downloads-apt-get) SDK; run `gcloud init`
2. Install psql client: `apt install postgresql-client-12`
3. Download and install the [SQL Proxy](https://cloud.google.com/sql/docs/postgres/connect-admin-proxy) (ignore steps 3,4,5 from the SQL Proxy guide)
4. Disconnect from VPN
5. Run step 4 to start the [SQL Proxy](https://cloud.google.com/sql/docs/postgres/connect-admin-proxy#unix-sockets)

    ```sh
    sudo mkdir /cloudsql; sudo chmod 777 /cloudsql
    ./cloud_sql_proxy -dir=/cloudsql &
    ```

6. Connect to your instance from the psql client(ex. psql -u test_user --host 127.0.0.1)

    `psql "sslmode=disable host=/cloudsql/<INSTANCE_CONNECTION_NAME> user=<USERNAME>"`
