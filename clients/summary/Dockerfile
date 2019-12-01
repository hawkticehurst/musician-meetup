# add the necessary instructions
# to create a Docker container image
# for your summary client using the
# NGINX web server image as the base

FROM nginx
COPY index.html /usr/share/nginx/html
COPY index.js /usr/share/nginx/html
COPY signup.html /usr/share/nginx/html
COPY signup.js /usr/share/nginx/html
COPY home.html /usr/share/nginx/html
COPY home.js /usr/share/nginx/html
COPY home.css /usr/share/nginx/html
ADD default.conf /etc/nginx/conf.d/default.conf

