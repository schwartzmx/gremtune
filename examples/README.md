# Examples

The provided exampled needs a germlin-server as counterpart. They can be either a local one or a cosmos db running on azure.

## Against Local Gremlin-Server

```bash
# start a local gremlin server
cd ..
make infra.up

# execute the example
go run ./examples/local
```

## Against CosmosDB

### Prerequisites

1. Obtain the host/ endpoint of the CosmosDB

   - Therefore one has to login into the Azure web console and navigate to the azure cosmos DB panel.
   - Here just select the db you want to connect to, go to the overview tab and there copy the URL of the `Gremlin Endpoint`
   - It should look like this `wss://xyz-my-db-abc.gremlin.cosmos.azure.com:443/`

2. Obtain the username/ database name

   - In our case the username is composed out of the name of the database and the name of the collection one want's to use
   - Schema: `/dbs/<db-name>/colls/<name of the collection>`
   - Example: `/dbs/mydatabase/colls/mygraph`

3. Obtain the key of the CosmosDB

   - Therefore one has to login into the Azure web console and navigate to the azure cosmos DB panel.
   - Here just select the db you want to connect to, go to the keys tab and there copy the `PRIMARY KEY`
   - It should look like this `Amx5obCQPF53Vf7WyqtXiu5qsZ89tS8envY9oON3KNRIGAjILWduRKbfwqvYZ2e8vtIUNNv1w0LlEoecfhNsk9w==`

4. Set these parameters as environment variables

   ```bash
   export CDB_HOST=wss://xyz-my-db-abc.gremlin.cosmos.azure.com:443 && \
   export  CDB_KEY=Amx5obCQPF53Vf7WyqtXiu5qsZ89tS8envY9oON3KNRIGAjILWduRKbfwqvYZ2e8vtIUNNv1w0Ll EoecfhNsk9w== && \
   export CDB_USERNAME=/dbs/mydatabase/colls/mygraph
   ```

Then one can run the example via:

```bash
cd ..
go run ./examples/cosmos
```
