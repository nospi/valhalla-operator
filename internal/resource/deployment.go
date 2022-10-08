package resource

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const DeploymentSuffix = ""

type DeploymentBuilder struct {
	*ValhallaResourceBuilder
}

func (builder *ValhallaResourceBuilder) Deployment() *DeploymentBuilder {
	return &DeploymentBuilder{builder}
}

func (builder *DeploymentBuilder) Build() (client.Object, error) {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      builder.Instance.ChildResourceName(DeploymentSuffix),
			Namespace: builder.Instance.Namespace,
		},
	}, nil
}

func (builder *DeploymentBuilder) Update(object client.Object) error {
	name := builder.Instance.ChildResourceName(DeploymentSuffix)
	deployment := object.(*appsv1.Deployment)
	// pbfFileName := builder.Instance.Spec.GetPbfFileName()

	deployment.Spec = appsv1.DeploymentSpec{
		Replicas: builder.Instance.Spec.MinReplicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"app": name,
			},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": name,
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  name,
						Image: builder.Instance.Spec.GetImage(),
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 5000,
							},
						},
						Resources: corev1.ResourceRequirements{
							Requests: map[corev1.ResourceName]resource.Quantity{
								"memory": resource.MustParse("1Gi"),
								"cpu":    resource.MustParse("1"),
							},
						},
						Command: []string{
							"/bin/sh",
							"-c",
						},
						Args: []string{
							fmt.Sprintf(`
								cd %s && \
								valhalla_service valhalla.json 2
							`, valhallaDataPath),
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      name,
								MountPath: valhallaDataPath,
							},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: name,
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: name,
								ReadOnly:  true,
							},
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(builder.Instance, deployment, builder.Scheme); err != nil {
		return fmt.Errorf("failed setting controller reference: %v", err)
	}

	return nil
}