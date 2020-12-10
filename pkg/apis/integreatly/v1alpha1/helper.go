package v1alpha1

func (g *Grafana) UsedPersistentVolume() bool {
	return g.Spec.DataStorage != nil
}
