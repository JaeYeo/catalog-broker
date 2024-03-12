package com.infranics.cloudcontainer;

import org.apache.commons.lang3.tuple.ImmutablePair;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.cloud.servicebroker.model.binding.*;
import org.springframework.cloud.servicebroker.service.ServiceInstanceBindingService;
import reactor.core.publisher.Mono;


import org.springframework.stereotype.Service;

import java.util.Map;

@Service
public class CloudContainerServiceBinding implements ServiceInstanceBindingService {


    @Autowired
    CloudContainerService service;

    @Override
    public Mono<CreateServiceInstanceBindingResponse> createServiceInstanceBinding(CreateServiceInstanceBindingRequest request) {
        String serviceInstanceId = request.getServiceInstanceId();
        String bindingId = request.getBindingId();

        //
        // create credentials and store for later retrieval
        //

        ImmutablePair<Boolean, Map<String, Object >> ret = service.createOrGetBinding(serviceInstanceId,bindingId);

        Boolean is_existed = ret.getKey();
        Map<String, Object > credentials = ret.getValue();

        CreateServiceInstanceBindingResponse response = CreateServiceInstanceAppBindingResponse.builder()
                .credentials(credentials)
                .async(false)
                .bindingExisted(is_existed)
                .build();

        return Mono.just(response);
    }

    @Override
    public Mono<DeleteServiceInstanceBindingResponse> deleteServiceInstanceBinding(DeleteServiceInstanceBindingRequest request) {
        String serviceInstanceId = request.getServiceInstanceId();
        String bindingId = request.getBindingId();

        //
        // delete any binding-specific credentials
        //

        service.deleteBinding(serviceInstanceId,bindingId);

        return Mono.just(DeleteServiceInstanceBindingResponse.builder()
                .async(false)
                .build());
    }

    @Override
    public Mono<GetServiceInstanceBindingResponse> getServiceInstanceBinding(GetServiceInstanceBindingRequest request) {
        String serviceInstanceId = request.getServiceInstanceId();
        String bindingId = request.getBindingId();

        //
        // retrieve the details of the specified service binding
        //

        ImmutablePair<Boolean, Map<String, Object >> ret = service.getBinding(serviceInstanceId,bindingId);

        Map<String, Object > credentials = ret.getValue();

        GetServiceInstanceBindingResponse response = GetServiceInstanceAppBindingResponse.builder()
                .credentials(credentials)
                .build();

        return Mono.just(response);
    }

}
