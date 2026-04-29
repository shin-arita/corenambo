FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive
ENV LANG=ja_JP.UTF-8
ENV TZ=Asia/Tokyo
ENV PGDATA=/var/lib/postgresql/data

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      ca-certificates \
      lsb-release \
      wget \
      gnupg \
      locales \
      tzdata \
      gosu && \
    locale-gen ja_JP.UTF-8 && \
    rm -rf /var/lib/apt/lists/*

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      postgresql-common && \
    /usr/share/postgresql-common/pgdg/apt.postgresql.org.sh -y

COPY ./groonga.gpg /usr/share/keyrings/groonga.gpg
RUN echo "deb [signed-by=/usr/share/keyrings/groonga.gpg] https://packages.groonga.org/ubuntu/ noble universe" > /etc/apt/sources.list.d/groonga.list

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      postgresql-16 \
      postgresql-client-16 \
      postgresql-16-pgdg-pgroonga \
      groonga-tokenizer-mecab && \
    rm -rf /var/lib/apt/lists/*

RUN mkdir -p /var/lib/postgresql && \
    chown -R postgres:postgres /var/lib/postgresql

COPY ./migrations/000000_init_user.sql /docker-entrypoint-initdb.d/000000_init_user.sql

COPY ./entrypoint.sh /usr/local/bin/entrypoint.sh

RUN chmod +x /usr/local/bin/entrypoint.sh

EXPOSE 5432

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["/usr/lib/postgresql/16/bin/postgres", "-D", "/var/lib/postgresql/data", "-c", "listen_addresses=*"]
