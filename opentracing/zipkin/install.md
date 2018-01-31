# zipkin

# install

## docker 

```bash
docker run -d -p 9411:9411 openzipkin/zipkin
```

## java

```bash
wget -O zipkin.jar 'https://search.maven.org/remote_content?g=io.zipkin.java&a=zipkin-server&v=LATEST&c=exec'
java -jar zipkin.jar
```

## 源码

```bash
# get the latest source
git clone https://github.com/openzipkin/zipkin
cd zipkin
# Build the server and also make its dependencies
./mvnw -DskipTests --also-make -pl zipkin-server clean install
# Run the server
java -jar ./zipkin-server/target/zipkin-server-*exec.jar

```

# 配置参数

开启划线器
SCRIBE_ENABLED=true