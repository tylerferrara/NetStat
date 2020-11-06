# BULD first
# npm run build

# pull official base image
FROM nginx:alpine

COPY ./build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

ENTRYPOINT [ "nginx", "-g", "daemon off;" ]