FROM tinkerpop/gremlin-server:3.4.0

COPY gremlin-server.yaml /opt/gremlin-server/conf/gremlin-server.yaml
COPY tinkergraph-empty.properties conf/tinkergraph-empty.properties
COPY initialize_graph.groovy scripts/initialize_graph.groovy

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["conf/gremlin-server.yaml"]
