package com.infranics.cloudcontainer;

import java.util.Map;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.cloud.servicebroker.model.instance.*;
import reactor.core.publisher.Mono;

import org.springframework.cloud.servicebroker.service.ServiceInstanceService;
import org.springframework.stereotype.Service;

@Service
public class CloudContainerServiceInstance implements ServiceInstanceService {

    @Autowired
    CloudContainerService service;

    @Override
    public Mono<CreateServiceInstanceResponse> createServiceInstance(CreateServiceInstanceRequest request) {
        String serviceInstanceId = request.getServiceInstanceId();
        String planId = request.getPlanId();
        String serviceId = request.getServiceDefinitionId();
        Map<String, Object> parameters = request.getParameters();


        //
        // perform the steps necessary to initiate the asynchronous
        // provisioning of all necessary resources
        //

        //String dashboardUrl = ""; /* construct a dashboard URL */

        // instances dir 생성 및 terraform으로 vm 생성
        boolean instance_existed = service.createInstance(serviceInstanceId, serviceId, planId, parameters);

        return Mono.just(CreateServiceInstanceResponse.builder()
                .dashboardUrl(null)
                .async(false)
                .instanceExisted(instance_existed)
                .build());

    }



    @Override
    public Mono<DeleteServiceInstanceResponse> deleteServiceInstance(DeleteServiceInstanceRequest request) {
        String serviceInstanceId = request.getServiceInstanceId();
        String planId = request.getPlanId();

        //
        // perform the steps necessary to initiate the asynchronous
        // deletion of all provisioned resources
        //

        // instances dir 삭제
        service.deleteInstance(serviceInstanceId);

        return Mono.just(DeleteServiceInstanceResponse.builder()
                .async(false)
                .build());
    }

    /*
    @Override
    public Mono<UpdateServiceInstanceResponse> updateServiceInstance(UpdateServiceInstanceRequest request) {
        String serviceInstanceId = request.getServiceInstanceId();
        String planId = request.getPlanId();
        String previousPlan = request.getPreviousValues().getPlanId();
        Map<String, Object> parameters = request.getParameters();

        //
        // perform the steps necessary to initiate the asynchronous
        // updating of all necessary resources
        //

        return Mono.just(UpdateServiceInstanceResponse.builder()
                .async(false)
                .build());
    }
    */

    @Override
    public Mono<GetServiceInstanceResponse> getServiceInstance(GetServiceInstanceRequest request) {
        String serviceInstanceId = request.getServiceInstanceId();

        //
        // retrieve the details of the specified service instance
        //

        String dashboardUrl = "";

        service.getInstance(serviceInstanceId);

        return Mono.just(GetServiceInstanceResponse.builder()
                .dashboardUrl(dashboardUrl)
                .build());
    }

    /*
    @Override
    public Mono<GetLastServiceOperationResponse> getLastOperation(GetLastServiceOperationRequest request) {
        String serviceInstanceId = request.getServiceInstanceId();

        //
        // determine the status of the operation in progress
        //

        return Mono.just(GetLastServiceOperationResponse.builder()
                .operationState(OperationState.FAILED)
                .build());
    }
    */
}
