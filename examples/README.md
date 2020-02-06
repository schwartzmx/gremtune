# Examples

The provided examples needs a gremlin-server as counterpart. The counterpart can either be a local one, e.g. gremlin-server, or a CosmosDB running on Azure.

## Against Local Gremlin-Server

```bash
# start a local gremlin server
cd ..
make infra.up

# execute the example
go run ./examples/local
```

## Against CosmosDB on Azure

### Prerequisites

1. Obtain the host/ endpoint of the CosmosDB

   - To this end one has to login into the Azure web console and navigate to the Azure CosmosDB panel.
   - Select the db you want to connect to, go to the overview tab and there copy the URL of the `Gremlin Endpoint`
   - It should look like this `wss://xyz-my-db-abc.gremlin.cosmos.azure.com:443/`

2. Obtain the username/ database name

   - In our case the username is composed of the database name and the name of the collection one want's to use
   - Schema: `/dbs/<db-name>/colls/<name of the collection>`
   - Example: `/dbs/mydatabase/colls/mygraph`

3. Obtain the key of the CosmosDB

   - Either, obtain a resource token (**recommended**)
     - See: [CosmosDB Resource Tokens](https://docs.microsoft.com/en-us/rest/api/cosmos-db/access-control-on-cosmosdb-resources#resource-tokens)
   - Or, obtain the master key (**not recommended**)

     - For this purpose one has to login into the Azure web console and navigate to the Azure CosmosDB panel.
     - Select the db you want to connect to, go to the keys tab and there copy the `PRIMARY KEY`
     - It should look like this `Amx5obCQPF53Vf7WyqtXiu5qsZ89tS8envY9oON3KNRIGAjILWduRKbfwqvYZ2e8vtIUNNv1w0LlEoecfhNsk9w==`

4. Set these parameters as environment variables

   ```bash
   export CDB_HOST=wss://xyz-my-db-abc.gremlin.cosmos.azure.com:443 && \
   export  CDB_KEY=Amx5obCQPF53Vf7WyqtXiu5qsZ89tS8envY9oON3KNRIGAjILWduRKbfwqvYZ2e8vtIUNNv1w0LlEoecfhNsk9w== && \
   export CDB_USERNAME=/dbs/mydatabase/colls/mygraph
   ```

Then one can run the example via:

```bash
cd ..
go run ./examples/cosmos
```
