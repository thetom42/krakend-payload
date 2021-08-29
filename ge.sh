#/bin/bash
docker stop ge >> /dev/null 2>&1
docker rm ge >> /dev/null 2>&1
docker build -t go116 -f Dockerfile .
docker run -it -v $PWD:/temp --name ge go116 build.sh
mv -f krakend-payload.so ../gateway-test/plugins