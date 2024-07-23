# Controller Runtime Exercise

This repo contains a learn-by-doing style exercise to familiarize users with the inner workings of the [controller runtime](https://github.com/kubernetes-sigs/controller-runtime). Many projects, like [Argo CD](https://github.com/argoproj/argo-cd), [Istio](https://github.com/istio/istio), and everything that leverages the [Operator Framework](https://operatorframework.io/). The most common use case is to think of Kubernetes resources (those little yaml definitions) as defining the intended state of the world, and services that leverage the Controller Runtime are actors that will "reconcile" the state of the world to the intended state.

## Level vs Edge Based Reconciliation

The Controller Runtime can leverage both level and edge based reconiliation. Edge based watches for events, like resource creation, deletion, or updates and reconciles the state of the world to match the intent. Level based periodically reads in the current intent, and reconciles so the state of the world matches the intent.

## Your Task

You are the head of your dev ops department, and you'd like to provide your engineers with a simpler interface for defining their services with a common set of K8s best practices. To do so, you plan to defin a new K8s resource `MyApp` that provides a smaller API surface, and a lot of defaults around resources, pod disruption budgets, and more.

1. Modify the `Reconcile` method to create a Deployment with the provided image, a default set of resources (CPU and mem), and a pod disruption budget.

1. Add some custom metrics to the metrics handler, and enable the metrics handler. What's an example metric you might want to keep track of?

1. Enable leader election

1. Use a custom queue that orders Reconciliation by name, lexicographically.

## Testing Your Understanding

Answer the following questions to test your understanding. You may want to tweak your code or add print/debug statements to test how it works.

1. What are the event types that the reconciler can recieve?

1. Which controller option controls how often Level-based reconciliation occurs? 

1. Which event types are passed in on level based reconciliation?

1. What is the difference between the `builder.Builder`'s `For`, `Owns` and `Watches` methods?

1. What happens if 10 updates to the same object occur in rapid succession (ie: before a single Reconcile occurs)? How many times is Reconcile called, and with which version of the object?

1. How can you control the speed of reconciliation?

1. How can you retry a failed reconiliation at a later time?

1. What happens to child objects if you delete a watched object?

1. Does the reconciler trigger on updates to the watched object, updates to the child object, or both?

1. Answer the above question, but for children of children? ie: 1 controller that creates a child object, that in turn creates a child object (ie: creating a deployment, will in turn create Pods)

1. What happens if a Reconciliation fails, and new updates come in?

1. What happens if a reconciliation fails to a child's child objects (ie: a deployments pods)?

1. If a single resource is in a failed state, does it block reconciliation of other objects?

1. A ReconcileRequest only has the `NamespacedName`. How do you get the full object? Is this object cached, or result in an API call to the k8s master?

1. What metrics does the controller runtime emit? Describe what some of those metrics represent

1. What is leader election, and when would you use it?