resource "openstack_compute_instance_v2" "tf_instance_01" {
  name            	= "${name}"
  flavor_id     	= "${flavor_id}"
  image_id      	= "${image_id}"
  key_pair        	= "${key_pair}"
  security_groups 	= ["${security_groups}"]

<#if initial_password?? >
  admin_pass        = "${initial_password}"
</#if>

  network {
    uuid = "${network_uuid}"
  }


  user_data       = <<-EOF
                      #cloud-config
                      password: "${initial_password}"
                      chpasswd: { expire False }
                      ssh_pwauth: True
                      timezone: Asia/Seoul
<#if    plan_name == "apache-ubuntu20.04-medium" ||
        plan_name == "nginx-ubuntu20.04-medium" ||
        plan_name == "tomcat-ubuntu20.04-medium" ||
        plan_name == "wildfly-ubuntu20.04-medium" ||
        plan_name == "ubuntu20.04-medium" ||
        plan_name == "centos8-medium" ||
        plan_name == "mysql-ubuntu20.04-medium" ||
        plan_name == "mariadb-ubuntu20.04-medium" ||
        plan_name == "postgresql-ubuntu20.04-medium" ||
        plan_name == "apache-ubuntu20.04-small" ||
        plan_name == "nginx-ubuntu20.04-small" ||
        plan_name == "tomcat-ubuntu20.04-small" ||
        plan_name == "wildfly-ubuntu20.04-small" ||
        plan_name == "ubuntu20.04-small" ||
        plan_name == "centos8-small" ||
        plan_name == "mysql-ubuntu20.04-small" ||
        plan_name == "mariadb-ubuntu20.04-small" ||
        plan_name == "postgresql-ubuntu20.04-small" >
                      runcmd:
                        - sudo sed -i 's/FILEBEAT_IP/${beat_ip}/g' /usr/share/filebeat/filebeat.yml
                        - sudo sed -i 's/FILEBEAT_PORT/${beat_port}/g' /usr/share/filebeat/filebeat.yml
                        - sudo systemctl daemon-reload
                        - sudo systemctl start filebeat
                        - sudo systemctl start node_exporter
</#if>
                    EOF

}

<#if floatingip_pool?? >

resource "openstack_compute_floatingip_v2" "fip_01" {
  pool = "${floatingip_pool}"
}

# 플로팅 IP 연결
resource "openstack_compute_floatingip_associate_v2" "fip_associate" {
  floating_ip = openstack_compute_floatingip_v2.fip_01.address
  instance_id = openstack_compute_instance_v2.tf_instance_01.id
}

</#if>





