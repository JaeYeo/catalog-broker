FROM openjdk:11-jdk-slim

RUN apt update && \
    apt install wget unzip -y 

WORKDIR /workspace

COPY ./target/cloudcontainer-0.0.1-SNAPSHOT.jar cloudcontainerbroker.jar 

RUN mkdir /workspace/data

ENTRYPOINT ["java", "-jar", "cloudcontainerbroker.jar" , "--cloudcontainerservicebroker.data_path=/workspace/data"]
