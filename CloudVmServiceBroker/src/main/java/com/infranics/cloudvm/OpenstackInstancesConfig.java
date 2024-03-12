package com.infranics.cloudvm;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Component
@ConfigurationProperties("cloudvmservicebroker")
public class OpenstackInstancesConfig {

    private List<Map<String, String>> openstack_instances;
    private HashMap<String, Map<String, String>> openstack_instance_hashmap;

    //
    // hashmap에서 openstack instance template 정보 찾아옴
    //
    public Map<String, String> findByKey(String key)
    {
        if(openstack_instance_hashmap==null)
        {
            // hashmap에 key,value로 저장
            openstack_instance_hashmap = new HashMap<String, Map<String, String>>();
            for (Map<String, String> template : openstack_instances) {
                String openstack_instance_key = getKey(template.get("service_id"), template.get("plan_id"));
                openstack_instance_hashmap.put(openstack_instance_key, template);
            }
        }

        return openstack_instance_hashmap.get(key);
    }
    static public String getKey(String service_id, String plan_id)
    {
        return service_id + ":" + plan_id;
    }

    // getter,setter
    public List<Map<String, String>> getOpenstack_instances() {

        return openstack_instances;
    }

    public void setOpenstack_instances(List<Map<String, String>> openstack_instances) {

        openstack_instance_hashmap = null;

        this.openstack_instances = openstack_instances;
    }

}