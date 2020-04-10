#!/usr/bin/env bash

PORT=33060
DB_NAME='apiserver_db'

PASS_LEN=20
NUMBYTES=`echo ${PASS_LEN} | awk '{print int($1*1.16)+1}'`
NEW_PASSWORD="$(openssl rand -base64 ${NUMBYTES} | tr -d "=+/" | cut -c1-${PASS_LEN})"

DOCKER_INSTANCE_NAME='apiserver_mysql'

ALREADY_RUNNING=`docker inspect -f '{{.State.Running}}' ${DOCKER_INSTANCE_NAME}`
if [[ ${ALREADY_RUNNING} = true ]]; then
    NEW_PASSWORD=`docker exec -it ${DOCKER_INSTANCE_NAME} sh -c "cat /root/mysql_pass.txt | tr -d '\n'"`
    echo "Container is already running with password ${NEW_PASSWORD}"
else
    docker pull mysql/mysql-server:8.0
    echo "Starting mysql with password ${NEW_PASSWORD}"
    docker run --name=${DOCKER_INSTANCE_NAME} -d -p ${PORT}:3306 -e MYSQL_ROOT_PASSWORD=${NEW_PASSWORD} -e MYSQL_DATABASE=${DB_NAME} -e MYSQL_ROOT_HOST=% mysql/mysql-server:8.0
    sleep 5
    READY=''
    while [[ -z ${READY} ]]; do
        READY=`docker logs ${DOCKER_INSTANCE_NAME} 2>&1 | grep '/usr/sbin/mysqld: ready for connections'`
        echo "Waiting for mysql to be ready"
        sleep 2
    done
    echo "Server ready:   ${READY}"
    docker exec -it ${DOCKER_INSTANCE_NAME} sh -c "echo ${NEW_PASSWORD} > /root/mysql_pass.txt"
    echo "Started mysql instance on localhost:${PORT} with root user and '$NEW_PASSWORD'"
fi

echo "Now run npm init with DATABASE_URI set to: "
echo "mysql://root:${NEW_PASSWORD}@127.0.0.1:${PORT}/${DB_NAME}"
echo "OR connect to mysql from host"
echo "mysql -h 127.0.0.1 -P ${PORT} -u root -p${NEW_PASSWORD}"
echo "OR connect to docker bash"
echo "docker exec -it ${DOCKER_INSTANCE_NAME} bash"
echo "OR connect to mysql inside docker"
echo "docker exec -it ${DOCKER_INSTANCE_NAME} sh -c 'mysql -p${NEW_PASSWORD}'"
