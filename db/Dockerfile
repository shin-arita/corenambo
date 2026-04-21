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

RUN wget https://packages.groonga.org/ubuntu/groonga-apt-source-latest-$(lsb_release --codename --short).deb && \
    apt-get update && \
    apt-get install -y --no-install-recommends ./groonga-apt-source-latest-$(lsb_release --codename --short).deb && \
    rm -f groonga-apt-source-latest-$(lsb_release --codename --short).deb

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      postgresql-16 \
      postgresql-client-16 \
      postgresql-16-pgroonga \
      groonga-tokenizer-mecab && \
    rm -rf /var/lib/apt/lists/*

RUN mkdir -p /var/lib/postgresql/data && \
    chown -R postgres:postgres /var/lib/postgresql

COPY ./entrypoint.sh /usr/local/bin/entrypoint.sh
COPY ./initdb /docker-entrypoint-initdb.d

RUN chmod +x /usr/local/bin/entrypoint.sh

EXPOSE 5432

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["/usr/lib/postgresql/16/bin/postgres", "-D", "/var/lib/postgresql/data"]