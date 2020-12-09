FROM nginx:alpine


ENV NGINX_HOST=localhost\
  NGINX_PORT=8080 \
  WORKER_PROCESSES=1\
  WORKER_CONNECTIONS=1024\
  KEEP_ALIVE_TO=65 \
  PROXY_ADD_X_FW=proxy_add_x_forwarded_for \
  PROXY_HOST=host \
  HTTP_UPGRADE=http_upgrade

COPY content /usr/share/nginx/html
COPY templates/nginx.tmpl /etc/nginx/nginx.tmpl

CMD /bin/sh -c "envsubst < /etc/nginx/nginx.tmpl > /etc/nginx/nginx.conf && nginx -g 'daemon off;' || cat /etc/nginx/nginx.conf"


  
