# Cosmos DB Response Error Handling

The status codes one can expect from Cosmos DB are listed at [HTTP Status Codes for Azure Cosmos DB](https://docs.microsoft.com/en-us/rest/api/cosmos-db/http-status-codes-for-cosmosdb).

But these status codes are not reported on at the layer of the gremlin api layer. Instead they are part of the received response data.
A gremlin message has the following structure:

```json
{
  "requestId": "string",
  "status": {
    "code": "int",
    "message": "string",
    "attributes": {
      "key": "value"
    }
  },
  "result": {
    "data": "string ([]byte)",
    "meta": {
      "key": "value"
    }
  }
}
```

For CosmosDB this structure is filled like this (error response for request rate limit exceeded):

```json
{
  "requestId": "cfe23609-abcd-efgh-ijkl-326cd091aa37",
  "status": {
    "code": 500,
    "message": "\r\n\nActivityId : 00000000-0000-0000-0000-000000000000\nExceptionType : DocumentClientException\nExceptionMessage ....",
    "attributes": {
      "x-ms-retry-after-ms": "00:00:09.0530000",
      "x-ms-substatus-code": 3200,
      "x-ms-source": "Microsoft.Azure.Documents.Client",
      "x-ms-status-code": 429,
      "x-ms-request-charge": 3779.34,
      "x-ms-total-request-charge": 3779.34,
      "x-ms-server-time-ms": 1056.2705,
      "x-ms-total-server-time-ms": 1056.2705,
      "x-ms-activity-id": "fdd08592-abcd-efgh-ijkl-97d35c2dda52"
    }
  },
  "result": {
    "data": null,
    "meta": {}
  }
}
```

As one can see:

1. The Status Code is 500 (Internal Server Error). The cause for that error is encoded in the attributes map as `"x-ms-status-code": 429`. This Cosmos DB specific status code means `Too Many Request`.
2. The entries in the attribute map represent the [Azure Cosmos DB Gremlin server response headers](https://docs.microsoft.com/en-us/azure/cosmos-db/gremlin-headers).
3. The attribute map contains information that is worth creating a metric for (e.g. `x-ms-request-charge`, `x-ms-server-time-ms`)
