FROM redis:7.0.5
WORKDIR /redis
COPY sentinel_entrypoint.sh /usr/local/bin/
EXPOSE 26379
ENTRYPOINT ["sentinel_entrypoint.sh"]
