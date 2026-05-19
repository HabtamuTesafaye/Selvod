FROM nginx:alpine

# The envsubst tool is used to inject SV_STREAM_SECRET into the Nginx config at runtime.
# Standard Nginx image has this built-in via /docker-entrypoint.d/20-envsubst-on-templates.sh

COPY default.conf.template /etc/nginx/templates/default.conf.template
