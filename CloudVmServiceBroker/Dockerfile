FROM openjdk:11-jdk-slim

RUN apt update && \
    apt install wget unzip -y 

RUN wget https://releases.hashicorp.com/terraform/1.3.7/terraform_1.3.7_linux_amd64.zip && \
    unzip terraform_1.3.7_linux_amd64.zip && \
    mv terraform /usr/local/bin/

WORKDIR /workspace

COPY ./target/cloudvm-0.0.1-SNAPSHOT.jar cloudvmbroker.jar 

RUN mkdir /workspace/data

ENTRYPOINT ["java", "-jar", "cloudvmbroker.jar" , "--cloudvmservicebroker.data_path=/workspace/data"]
