# Local Gremlin Server

## Gremlin Console

### Installation

Since CosmosDB supports only [Tinkerpop 3.4.0](https://tinkerpop.apache.org/docs/3.4.0) one have to download the [gremlin console in Version 3.4.0](https://archive.apache.org/dist/tinkerpop/3.4.0/apache-tinkerpop-gremlin-console-3.4.0-bin.zip).

```bash
wget https://archive.apache.org/dist/tinkerpop/3.4.0/apache-tinkerpop-gremlin-console-3.4.0-bin.zip
unzip apache-tinkerpop-gremlin-console-3.4.0-bin.zip
sudo mkdir -p /opt/tinkerpop/gremlin
sudo mv apache-tinkerpop-gremlin-console-3.4.0 /opt/tinkerpop/gremlin
sudo ln -s /opt/tinkerpop/gremlin/apache-tinkerpop-gremlin-console-3.4.0/bin/gremlin.sh /usr/bin/gremlin

# then check if it is working
gremlin --version
# output: gremlin 3.4.0
```

### Connect to Gremlin Server

```bash
:remote connect tinkerpop.server conf/remote.yaml

# issue one command using the remote connection
:>

# e.g.
:> g.V()

# switch the console to execute all commands on the remote server
:remote console
```
