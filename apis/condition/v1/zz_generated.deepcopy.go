package v1

func (in *StatusConditions) DeepCopyInto(out *StatusConditions) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]StatusCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

func (in *StatusCondition) DeepCopyInto(out *StatusCondition) {
	*out = *in
}
