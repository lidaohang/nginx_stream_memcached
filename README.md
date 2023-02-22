# nginx_stream_memcached

支持nginx tcp proxy代理  memcached实例，目前实现了一致性hash.

支持的指令：
<li> cmd  set </li>
<li> cmd  get </li>
<li> cmd  getrange </li>
<li> cmd  add </li>
<li> cmd  delete </li>


<h6> memcached启动指令： </h6>
<li>
 /usr/local/bin/memcached -u root -p 11211 -o ext_page_size=256,ext_path=/data1/memcached/data_file:1700G,ext_path=/data5/memcached/data_file:1700G,ext_item_age=10,ext_wbuf_size=32,ext_threads=2,hashpower=30 -t 16 -c 50000 -m 70000 -f 2.0 -n 1024 -B auto -I 8m -F
</li>
