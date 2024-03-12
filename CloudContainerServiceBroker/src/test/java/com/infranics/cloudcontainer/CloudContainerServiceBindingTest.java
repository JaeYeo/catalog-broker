package com.infranics.cloudcontainer;

import org.apache.commons.lang3.tuple.ImmutablePair;
import org.junit.jupiter.api.Test;

import java.util.Map;

public class CloudContainerServiceBindingTest {


    @Test
    public void bind_json_test()
    {
        /*
        CloudContainerService service = new CloudContainerService();

        service.data_path = "D:\\Work\\CloudContainerServiceBroker\\CloudContainerServiceBroker\\data";


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


         */

    }

    @Test
    public void unbind_json_test()
    {
        /*
        CloudContainerService service = new CloudContainerService();

        service.data_path = "D:\\Work\\CloudContainerServiceBroker\\CloudContainerServiceBroker\\data";

        try {
            service.deleteBinding("test", "test_binding");
        }
        catch(Exception e)
        {
            System.out.println(e);
        }

         */
    }


}
