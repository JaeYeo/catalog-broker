package com.infranics.cloudcontainer;

import com.fasterxml.jackson.databind.ObjectMapper;
import io.fabric8.kubernetes.api.model.HasMetadata;
import io.fabric8.kubernetes.api.model.Namespace;
import io.fabric8.kubernetes.api.model.NamespaceList;
import io.fabric8.kubernetes.api.model.ObjectMeta;
import io.fabric8.kubernetes.client.*;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.lang3.tuple.ImmutablePair;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.cloud.servicebroker.exception.*;
import org.springframework.cloud.servicebroker.model.instance.DeleteServiceInstanceResponse;
import org.springframework.stereotype.Component;
import org.springframework.util.FileSystemUtils;

import java.io.*;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.*;

@Slf4j
@Component("cloud_container_service")
public class CloudContainerService {

    @Value("${cloudcontainerservicebroker.data_path}")
    protected String data_path;

    @Autowired
    private ShellCommandExecutor cmd;

    public boolean createInstance(String serviceInstanceId, String serviceId, String planId, Map<String, Object> parameters) {

        log.info("create data_path:" + data_path + ",instanceId:" + serviceInstanceId);
        Path dir_path = Paths.get(data_path + (isWindow()?"\\":"/") + serviceInstanceId);

        boolean instance_existed = false;
        
        // 해당 namespace에 serviceInstanceId로 생성된 deployment가 있는지 확인

        try {

            if (Files.exists(dir_path)) {

                // kubectl get 실재로 존재하는지 확인 필요
                // meta.txt 정보
                Path metadata_path = Paths.get(data_path + (isWindow() ? "\\" : "/") + serviceInstanceId + (isWindow() ? "\\" : "/") + "metadata.txt");
                ObjectMapper mapper = new ObjectMapper();
                HashMap<String, Object> metadata = mapper.readValue(metadata_path.toFile(), HashMap.class);
                String namespace = metadata.get("namespace").toString();
                String kube_config = metadata.get("kube_config").toString();
                String deployment_yaml = metadata.get("deployment_yaml").toString();

                // 파라메터 정보
                String namespace_param = parameters.get("namespace").toString();
                String kube_config_param = parameters.get("kube_config").toString();
                String deployment_yaml_param = parameters.get("deployment_yaml").toString();

                if(namespace.equals(namespace_param) && kube_config.equals(kube_config_param) && deployment_yaml.equals(deployment_yaml_param)) {
                    // 200 ok, nothing to do
                    instance_existed = true;
                }
                else {
                    // error other instance exists
                    throw new ServiceInstanceExistsException(serviceInstanceId,serviceId);
                }

            } else {


                // create dir for instance
                Files.createDirectory(dir_path);

                // param에서 kubeconfg 가져오기

                String kube_config_param = parameters.get("kube_config").toString();
                byte[] decodedBytes = Base64.getDecoder().decode(kube_config_param);
                String kube_config = new String(decodedBytes, StandardCharsets.UTF_8);

                System.out.println(kube_config);

                String deployment_yaml_param = parameters.get("deployment_yaml").toString();
                decodedBytes = Base64.getDecoder().decode(deployment_yaml_param);
                String deployment_yaml = new String(decodedBytes, StandardCharsets.UTF_8);

                System.out.println(deployment_yaml);

                String namespace_param = parameters.get("namespace").toString();

                // kube config
                // write kube_config
                Path kubeconfig_path = Paths.get(data_path + (isWindow()?"\\":"/") + serviceInstanceId + (isWindow()?"\\":"/") + "kube_config");
                FileWriter kubeconfig_path_file_writer = new FileWriter(kubeconfig_path.toFile());
                kubeconfig_path_file_writer.write(kube_config);
                kubeconfig_path_file_writer.close();

                // deployment.yaml
                // write deployment.yaml
                Path deployment_path = Paths.get(data_path + (isWindow()?"\\":"/") + serviceInstanceId + (isWindow()?"\\":"/") + "deployment.yaml");
                FileWriter deployment_path_file_writer = new FileWriter(deployment_path.toFile());
                deployment_path_file_writer.write(deployment_yaml);
                deployment_path_file_writer.close();

                // kubectl create -f deployment.yaml -n namespace
                Config config = Config.fromKubeconfig(kube_config);
                KubernetesClient client = new KubernetesClientBuilder().withConfig(config).build();

                client.load(new ByteArrayInputStream(deployment_yaml.getBytes())).inNamespace(namespace_param).create();
                
                // meta.txt 저장
                Path metadata_path = Paths.get(data_path + (isWindow() ? "\\" : "/") + serviceInstanceId + (isWindow() ? "\\" : "/") + "metadata.txt");
                ObjectMapper mapper = new ObjectMapper();
                Map<String, Object> metadata = new HashMap<String, Object>();
                metadata.put("namespace",namespace_param);
                metadata.put("deployment_yaml",deployment_yaml_param);
                metadata.put("kube_config",kube_config_param);
                mapper.writeValue(metadata_path.toFile(), metadata);



                // 201 created
                instance_existed = false;
            }
        }
        catch(ServiceBrokerException sbe)  // service borker exception은 상위로 올리기
        {

            log.error(sbe.toString());

            try {deleteInstance(serviceInstanceId);}catch(Exception es) {}

            throw sbe;
        }
        catch(Exception e)  // internal error로 올리기
        {
            log.error(e.toString());

            try {deleteInstance(serviceInstanceId);}catch(Exception es) {}

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
            //
            // kubectl delete
            //

            // namespace 정보
            Path metadata_path = Paths.get(data_path + (isWindow() ? "\\" : "/") + serviceInstanceId + (isWindow() ? "\\" : "/") + "metadata.txt");
            ObjectMapper mapper = new ObjectMapper();
            HashMap<String, Object> metadata = mapper.readValue(metadata_path.toFile(), HashMap.class);
            String namespace = metadata.get("namespace").toString();

            // deployment.yaml 정보
            Path deployment_path = Paths.get(data_path + (isWindow() ? "\\" : "/") + serviceInstanceId + (isWindow() ? "\\" : "/") + "deployment.yaml");
            String deployment_yaml = new String(Files.readAllBytes(deployment_path),StandardCharsets.UTF_8);

            // kube_config 정보
            Path kubeconfig_path = Paths.get(data_path + (isWindow()?"\\":"/") + serviceInstanceId + (isWindow()?"\\":"/") + "kube_config");
            String kube_config = new String(Files.readAllBytes(kubeconfig_path),StandardCharsets.UTF_8);

            System.out.println(kube_config);

            Config config = Config.fromKubeconfig(kube_config);
            KubernetesClient client = new KubernetesClientBuilder().withConfig(config).build();

            //Config config = new ConfigBuilder().withFile(kubeconfig_path.toFile()).build();
            //KubernetesClient client = new KubernetesClientBuilder().withConfig(config).build();

            // kubectl delete -f deployment.yaml -n namespace
            client.load(new ByteArrayInputStream(deployment_yaml.getBytes())).inNamespace(namespace).delete();
        }
        catch(Exception e)
        {
            log.error(e.toString());
        }

        try {
            //
            // delete dir
            //
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

        try {
            //
            // kubectl get
            //

            // namespace 정보
            Path metadata_path = Paths.get(data_path + (isWindow() ? "\\" : "/") + serviceInstanceId + (isWindow() ? "\\" : "/") + "metadata.txt");
            ObjectMapper mapper = new ObjectMapper();
            HashMap<String, Object> metadata = mapper.readValue(metadata_path.toFile(), HashMap.class);
            String namespace = metadata.get("namespace").toString();

            // deployment.yaml 정보
            Path deployment_path = Paths.get(data_path + (isWindow() ? "\\" : "/") + serviceInstanceId + (isWindow() ? "\\" : "/") + "deployment.yaml");
            String deployment_yaml = new String(Files.readAllBytes(deployment_path),StandardCharsets.UTF_8);

            System.out.println(deployment_yaml);

            // kube_config 정보
            Path kubeconfig_path = Paths.get(data_path + (isWindow()?"\\":"/") + serviceInstanceId + (isWindow()?"\\":"/") + "kube_config");
            String kube_config = new String(Files.readAllBytes(kubeconfig_path),StandardCharsets.UTF_8);

            System.out.println(kube_config);

            Config config = Config.fromKubeconfig(kube_config);
            KubernetesClient client = new KubernetesClientBuilder().withConfig(config).build();

            //Config config = new ConfigBuilder().withFile(kubeconfig_path.toFile()).build();
            //KubernetesClient client = new KubernetesClientBuilder().withConfig(config).build();

            // kubectl get -f deployment.yaml -n namespace
            List<HasMetadata> response_list = client.load(new ByteArrayInputStream(deployment_yaml.getBytes())).inNamespace(namespace).get();

            for (HasMetadata item: response_list) {

                if(item==null)
                {
                    //
                    // delete dir 할지 판단 필요
                    //

                    throw new ServiceInstanceDoesNotExistException("The service instances of "+serviceInstanceId+" is incomplete;delete first");
                }
            }

        }
        catch(Exception e)
        {
            log.error(e.toString());
            throw new ServiceInstanceDoesNotExistException("No service instances of "+serviceInstanceId);
        }


    }


    // return a pair of is_created,credentails
    public ImmutablePair<Boolean,Map<String,Object>> createOrGetBinding(String serviceInstanceId,String bindingId)
    {
        /*
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
         */

        return null;
    }

    public ImmutablePair<Boolean,Map<String,Object>> getBinding(String serviceInstanceId,String bindingId)
    {

        /*
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

         */
        return null;
    }


    public void deleteBinding(String serviceInstanceId,String bindingId)
    {
        /*
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
        */

    }

    /**
     * is window?
     */
    private boolean isWindow() {
        return System.getProperty("os.name").toLowerCase().startsWith("windows");
    }

}
