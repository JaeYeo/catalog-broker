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

<#if    plan_name == "apache-ubuntu20.04-medium" ||
        plan_name == "nginx-ubuntu20.04-medium" ||
        plan_name == "tomcat-ubuntu20.04-medium" ||
        plan_name == "wildfly-ubuntu20.04-medium">
  user_data       = <<-EOF
                      #cloud-config
                      password: "${initial_password}"
                      chpasswd: { expire False }
                      ssh_pwauth: True
                      timezone: Asia/Seoul
                      runcmd:
                        - echo export beatip='${beat_ip}' >> /root/.profile
                        - echo export beatport='${beat_port}' >> /root/.profile
                        - source /root/.profile
                        - cd /root/filebeat-7.9.1-linux-x86_64
                        - ./filebeat-start.sh
                        - cd /root/node_exporter
                        - ./node_exporter_exec.sh
                    EOF
</#if>
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





