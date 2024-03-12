package com.infranics.cloudvm;

import org.apache.commons.lang3.tuple.ImmutablePair;
import org.junit.jupiter.api.Test;

import java.util.Map;

public class CloudVmServiceBindingTest {


    @Test
    public void bind_json_test()
    {
        CloudVmService service = new CloudVmService();

        service.data_path = "D:\\Work\\CloudVmServiceBroker\\CloudVmServiceBroker\\data";


        try {


            ImmutablePair<Boolean, Map<String, Object>> ret = service.createOrGetBinding("test", "test_binding");

            if (ret != null) {
                Boolean is_existed = ret.getKey();
                Map<String, Object> credentials = ret.getValue();

                System.out.println("test_binding_binding.txt is " + is_existed);
                System.out.println(credentials);
            }

        }
        catch(Exception e)
        {
            System.out.println(e);
        }


    }

    @Test
    public void unbind_json_test()
    {
        CloudVmService service = new CloudVmService();

        service.data_path = "D:\\Work\\CloudVmServiceBroker\\CloudVmServiceBroker\\data";

        try {
            service.deleteBinding("test", "test_binding");
        }
        catch(Exception e)
        {
            System.out.println(e);
        }
    }


}
