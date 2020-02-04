FROM tinkerpop/gremlin-server:3.4.0

COPY gremlin-server.yaml /opt/gremlin-server/conf/gremlin-server.yaml

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["conf/gremlin-server.yaml"]
