package v1alpha1

func (g *Grafana) UsedPersistentVolume() bool {
	return g.Spec.DataStorage != nil
}

func (i *Loki) UsedPersistentVolume() bool {
	return i.Spec.DataStorage != nil
}