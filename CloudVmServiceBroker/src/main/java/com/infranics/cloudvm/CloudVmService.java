package com.infranics.cloudvm;

import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.lang3.tuple.ImmutablePair;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.cloud.servicebroker.exception.ServiceBrokerException;
import org.springframework.cloud.servicebroker.exception.ServiceBrokerInvalidParametersException;
import org.springframework.cloud.servicebroker.exception.ServiceInstanceBindingDoesNotExistException;
import org.springframework.cloud.servicebroker.exception.ServiceInstanceDoesNotExistException;
import org.springframework.cloud.servicebroker.model.instance.DeleteServiceInstanceResponse;
import org.springframework.stereotype.Component;
import org.springframework.util.FileSystemUtils;

import java.io.*;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.Map;

@Slf4j
@Component("cloud_vm_service")
public class CloudVmService {

    @Value("${cloudvmservicebroker.data_path}")
    protected String data_path;

    @Value("${cloudvmservicebroker.beat_ip}")
    protected String beat_ip;
    @Value("${cloudvmservicebroker.beat_port}")
    protected String beat_port;

    @Autowired
    private OpenstackInstancesConfig openstack_instance_config;

    @Autowired
    private freemarker.template.Configuration freemarker_config;


    @Autowired
    private ShellCommandExecutor cmd;

    public boolean createInstance(String serviceInstanceId, String serviceId, String planId, Map<String, Object> parameters) {

        log.info("create data_path:" + data_path + ",instanceId:" + serviceInstanceId);

        System.out.println(openstack_instance_config.getOpenstack_instances());

        Path dir_path = Paths.get(data_path + (isWindow()?"\\":"/") + serviceInstanceId);

        boolean instance_existed = false;
        FileWriter provider_file_writer = null;
        FileWriter instance_file_writer = null;

        // check if directory exists
        try {


            if (Files.exists(dir_path)) {
                // 200 ok, nothing to do
                instance_existed = true;
            } else {
                // 201 created
                instance_existed = false;

                // create dir for instance
                Files.createDirectory(dir_path);

                // terraform config
                // create provider.tf of openstack
                Path provider_path = Paths.get(data_path + (isWindow()?"\\":"/") + serviceInstanceId + (isWindow()?"\\":"/") + "provider.tf");
                provider_file_writer = new FileWriter(provider_path.toFile());

                // generate provider.tf form template
                freemarker.template.Template freemarker_provider_template = freemarker_config.getTemplate("provider.ftl");

                Map<String,String> provider = new HashMap<>();

                // param으로 openstack 정보 얻어옴
                if(isOpenstackParametersValid(parameters)==false)
                    throw new ServiceBrokerInvalidParametersException("not enough openstack credentials");

                provider.put("user_name",parameters.get("user_name").toString());
                provider.put("tenant_name",parameters.get("tenant_name").toString());
                provider.put("domain_name",parameters.get("domain_name").toString());
                provider.put("password",parameters.get("password").toString());
                provider.put("auth_url",parameters.get("auth_url").toString());
                provider.put("region",parameters.get("region").toString());

                freemarker_provider_template.process(provider,provider_file_writer);

                provider_file_writer.close();

                // create instance.tf of service instances
                Path instance_path = Paths.get(data_path + (isWindow()?"\\":"/") + serviceInstanceId + (isWindow()?"\\":"/") + "instance.tf");
                instance_file_writer = new FileWriter(instance_path.toFile());

                freemarker.template.Template freemarker_instance_template = freemarker_config.getTemplate("instance.ftl");

                String instance_key = OpenstackInstancesConfig.getKey(serviceId, planId);

                System.out.println(instance_key);

                // other parameters already in the map found
                Map<String,String> instance = openstack_instance_config.findByKey(instance_key);

                System.out.println(instance);

                instance.put("name",serviceInstanceId);
                instance.put("key_pair",parameters.get("key_pair").toString());
                instance.put("network_uuid",parameters.get("network_uuid").toString());
                instance.put("security_groups",parameters.get("security_groups").toString());
                instance.put("beat_ip",beat_ip);
                instance.put("beat_port",beat_port);

                if(parameters.get("initial_password") == null)
                    instance.put("initial_password","password");
                else
                    instance.put("initial_password",parameters.get("initial_password").toString());


                freemarker_instance_template.process(instance,instance_file_writer);

                instance_file_writer.close();


                // terraform init
                cmd.exec(dir_path,"terraform init");

                // run terraform apply
                cmd.exec(dir_path,"terraform apply --auto-approve");

                //
                // check result if possible,
                //



            }
        }
        catch(ServiceBrokerException sbe)  // service borker exception은 상위로 올리기
        {
            try {
                if (provider_file_writer != null) provider_file_writer.close();
                if (instance_file_writer != null) instance_file_writer.close();
            } catch(Exception ie) {}

            deleteInstance(serviceInstanceId);

            throw sbe;
        }
        catch(Exception e)  // internal error로 올리기
        {
            try {
                if (provider_file_writer != null) provider_file_writer.close();
                if (instance_file_writer != null) instance_file_writer.close();
            } catch(Exception ie) {}

            log.error(e.toString());

            deleteInstance(serviceInstanceId);

            throw new ServiceBrokerException("500","Failed to create service instance",e);
        }

        return instance_existed;

    }



    public void deleteInstance(String serviceInstanceId)
    {

        log.info("delete data_path:" + data_path + ",instanceId:" + serviceInstanceId);

        Path dir_path = Paths.get(data_path + (isWindow()?"\\":"/") + serviceInstanceId);

        if (!Files.exists(dir_path))
        {
            throw new ServiceInstanceDoesNotExistException("No service instances of "+serviceInstanceId);
        }

        try {

            // destroy terraform
            cmd.exec(dir_path,"terraform destroy --auto-approve");

            // delete dir
            FileSystemUtils.deleteRecursively(dir_path);
        }
        catch(Exception e)
        {
            log.error(e.toString());
        }

    }

    public void getInstance(String serviceInstanceId)
    {

        log.info("get data_path:" + data_path + ",instanceId:" + serviceInstanceId);

        Path dir_path = Paths.get(data_path + (isWindow()?"\\":"/") + serviceInstanceId);

        if (!Files.exists(dir_path))
        {
            throw new ServiceInstanceDoesNotExistException("No service instances of "+serviceInstanceId);
        }

    }


    // return a pair of is_created,credentails
    public ImmutablePair<Boolean,Map<String,Object>> createOrGetBinding(String serviceInstanceId,String bindingId)
    {

        log.info("bind data_path:" + data_path + ",instanceId:" + serviceInstanceId);

        Path dir_path = Paths.get(data_path + (isWindow()?"\\":"/") + serviceInstanceId);
        if (!Files.exists(dir_path))
        {
            throw new ServiceInstanceDoesNotExistException("No service instances of "+serviceInstanceId);
        }


        ImmutablePair<Boolean,Map<String,Object>> ret = null ;

        try {

            // check create binding meta file
            Path binding_path = Paths.get(data_path + (isWindow() ? "\\" : "/") + serviceInstanceId + (isWindow() ? "\\" : "/") + bindingId + "_binding.txt");

            if (Files.exists(binding_path)) {
                ObjectMapper mapper = new ObjectMapper();

                HashMap<String, Object> credentials = mapper.readValue(binding_path.toFile(), HashMap.class);

                ret = new ImmutablePair<Boolean, Map<String, Object>>(true, credentials);
            } else {
                // get binding info from terraform.tfstate

                Map<String, Object> credentials = new HashMap<String, Object>();

                Path tfstate_path = Paths.get(data_path + (isWindow() ? "\\" : "/") + serviceInstanceId + (isWindow() ? "\\" : "/") + "terraform.tfstate");

                ObjectMapper mapper = new ObjectMapper();

                HashMap<String, Object> tfstat = mapper.readValue(tfstate_path.toFile(), HashMap.class);

                ArrayList<HashMap<String, Object>> resources = (ArrayList<HashMap<String, Object>>) tfstat.get("resources");

                HashMap<String, HashMap<String, Object>> resources_map = new HashMap<>();
                for (HashMap<String, Object> resource : resources) {
                    resources_map.put((String) resource.get("type"), resource);
                }

                HashMap<String, Object> floatingip_map = resources_map.get("openstack_compute_floatingip_associate_v2");
                if (floatingip_map != null) {

                    Map<String, Object> instance = (Map<String, Object>) ((ArrayList) floatingip_map.get("instances")).get(0);

                    String floating_ip = ((Map<String, String>) instance.get("attributes")).get("floating_ip");

                    credentials.put("FLOATING_IP", floating_ip);

                    //System.out.println("floating_ip="+floating_ip);

                }


                HashMap<String, Object> instance_map = resources_map.get("openstack_compute_instance_v2");
                if (instance_map != null) {

                    Map<String, Object> instance = (Map<String, Object>) ((ArrayList) instance_map.get("instances")).get(0);

                    String access_ip = ((Map<String, String>) instance.get("attributes")).get("access_ip_v4");
                    if(access_ip != null)
                        credentials.put("ACCESS_IP", access_ip);

                    String initial_password = ((Map<String, String>) instance.get("attributes")).get("admin_pass");
                    if(initial_password!=null)
                        credentials.put("INITIAL_PASSWORD", initial_password);

                    //System.out.println("access_ip="+access_ip);
                }

                // make create binding meta file
                mapper.writeValue(binding_path.toFile(), credentials);

                ret = new ImmutablePair<Boolean, Map<String, Object>>(false, credentials);

            }

        }
        catch(Exception e)
        {
            Map<String, Object> credentials = new HashMap<String, Object>();

            credentials.put("ERROR",e.toString());

            ret = new ImmutablePair<Boolean, Map<String, Object>>(false, credentials);

            log.error(e.toString());
        }

        return ret;
    }

    public ImmutablePair<Boolean,Map<String,Object>> getBinding(String serviceInstanceId,String bindingId)
    {

        log.info("get-bind data_path:" + data_path + ",instanceId:" + serviceInstanceId + ",bindingId:" + bindingId);

        Path binding_path = Paths.get(data_path + (isWindow()?"\\":"/") + serviceInstanceId + (isWindow()?"\\":"/") + bindingId + "_binding.txt");
        if (!Files.exists(binding_path))
        {
            throw new ServiceInstanceBindingDoesNotExistException("No binding of "+bindingId);
        }

        ImmutablePair<Boolean,Map<String,Object>> ret = null ;

        try
        {
            ObjectMapper mapper = new ObjectMapper();

            HashMap<String, Object> credentials = mapper.readValue(binding_path.toFile(), HashMap.class);

            ret = new ImmutablePair<Boolean, Map<String, Object>>(true, credentials);
        }
        catch(Exception e)
        {
            Map<String, Object> credentials = new HashMap<String, Object>();

            credentials.put("ERROR",e.toString());

            ret = new ImmutablePair<Boolean, Map<String, Object>>(false, credentials);

            log.error(e.toString());
        }

        return ret;
    }


    public void deleteBinding(String serviceInstanceId,String bindingId)
    {
        log.info("unbind data_path:" + data_path + ",instanceId:" + serviceInstanceId + ",bindingId:" + bindingId);

        Path binding_path = Paths.get(data_path + (isWindow()?"\\":"/") + serviceInstanceId + (isWindow()?"\\":"/") + bindingId + "_binding.txt");
        if (!Files.exists(binding_path))
        {
            throw new ServiceInstanceBindingDoesNotExistException("No binding of "+bindingId);
        }

        try {
            Files.deleteIfExists(binding_path);
        }
        catch(Exception e)
        {
            log.error(e.toString());
        }

    }

    /**
     * is window?
     */
    private boolean isWindow() {
        return System.getProperty("os.name").toLowerCase().startsWith("windows");
    }

    private boolean isOpenstackParametersValid(Map<String, Object> parameters) {

        if ( parameters != null &&
                parameters.get("user_name") != null &&
                parameters.get("tenant_name") != null &&
                parameters.get("domain_name") != null &&
                parameters.get("password") != null &&
                parameters.get("auth_url") != null &&
                parameters.get("region") != null &&
                parameters.get("key_pair") != null &&
                parameters.get("network_uuid") != null &&
                parameters.get("security_groups") != null
                )
            return true;
        else
            return false;
    }

}