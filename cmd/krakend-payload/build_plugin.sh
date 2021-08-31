#/bin/bash
docker stop compiler >> /dev/null 2>&1
docker rm compiler >> /dev/null 2>&1
docker build -t go116 -f Dockerfile.build .
docker run -it -v $PWD:/temp --name compiler go116 compile.sh
mv -f krakend-payload.so ../../dev/krakend/plugins
