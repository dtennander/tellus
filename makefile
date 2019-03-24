
IMAGE_TAG = $(shell find . -type f \( -exec sha1sum "$$PWD"/{} \; \) | awk '{print $$1}' | sort | sha1sum | cut -d ' ' -f 1 | cut -c -12)
IMAGE_REPO = tennander/tellus
IMAGE_NAME = $(IMAGE_REPO):$(IMAGE_TAG)

.PHONY: run

run: deploy logs

logs:
	while kubectl get all | grep pod/tellus-deployment | grep -q 0/1; do sleep 1; done
	kubectl logs -f deployment.apps/tellus-deployment

clean:
	kubectl delete -f k8s.yml

image: Dockerfile main.go tellus go.mod go.sum
	docker build . -t $(IMAGE_NAME)

push: image
	docker push $(IMAGE_NAME)

deploy: push k8s.yml
	sed "s#$(IMAGE_REPO)#$(IMAGE_NAME)#g" < k8s.yml >> k8s.yml.remove_me

deploy_locally: image k8s.yml
	sed "s#$(IMAGE_REPO)#$(IMAGE_NAME)#g;\
		 s#imagePullPolicy: IfNotPresent#imagePullPolicy: Never#g" < k8s.yml >> k8s.yml.remove_me
	kubectl apply -f k8s.yml.remove_me
	rm k8s.yml.remove_me

get_tag:
	@echo $(IMAGE_TAG)

get_image_name:
	@echo $(IMAGE_NAME)