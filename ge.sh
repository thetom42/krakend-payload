#/bin/bash
docker rm ge >> /dev/null 2>&1
docker run -it -v $PWD:/temp --name ge go116 build.sh
mv krakend-payload.so ../gateway-test/plugins