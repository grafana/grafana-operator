{
  // Prometheus datasource type
  PROMETHEUS_DS_TYPE:
    'prometheus',

  // Prometheus datasource uid
  PROMETHEUS_DS_UID:
    std.extVar('PROMETHEUS_DS_UID'),

  //Target namespace where service is deployed
  K8S_NAMESPACE:
    std.extVar('K8S_NAMESPACE')
}
