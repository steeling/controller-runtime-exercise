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

1. Modify the reconciler to ignore `MyApp`'s with labels `reconciler: ignore`

1. When a `MyApp` is deleted, remove all children.

1. Make sure the status of MyApp is kept up to date.

1. Migreate the controller runtime usage to leverage the [KubeBuilder](https://github.com/kubernetes-sigs/kubebuilder-declarative-pattern)

## Testing Your Understanding

Answer the following questions to test your understanding. You may want to tweak your code or add print/debug statements to test how it works.

1. What are the event types that the reconciler can recieve: Controller Request

1. Which controller option controls how often Level-based reconciliation occurs: Resync after

1. Which event types are passed in on level based reconciliation: Update, Create, Delete

1. What is the difference between the `builder.Builder`'s `For`, `Owns` and `Watches` methods: For is for which resource kind is used i.e MyApps, Owns: the MyApp will own the CRs instances, 

1. What happens if 10 updates to the same object occur in rapid succession (ie: before a single Reconcile occurs)? How many times is Reconcile called, and with which version of the object?: The events will be added to the reconciliation queue and will be handled in order for the atest version of the object

1. How can you control the speed of reconciliation: Requeue after

1. How can you retry a failed reconiliation at a later time: Requeue after

1. What happens to child objects if you delete a watched object: the child will also be delted

1. Does the reconciler trigger on updates to the watched object, updates to the child object, or both? both

1. Answer the above question, but for children of children? ie: 1 controller that creates a child object, that in turn creates a child object (ie: creating a deployment, will in turn create Pods) the child will be delted but not the child of the child. The reconciler doesn't trigger on the updates for children of children. If there are ny updates that conflict with the intent they will be handled in the next reconciliation cycle

1. What happens if a Reconciliation fails, and new updates come in? the update during which the error occured would be requeued for reconciliation. In any case the controller will eventully bring the resource to the desired state

1. What happens if a reconciliation fails to a child's child objects (ie: a deployments pods)? the controller will retry / we see that myapp-sample-2 pods keep getting restarted because of their low time span and kubernetes thinks they're crashing

1. If a single resource is in a failed state, does it block reconciliation of other objects? No

1. A ReconcileRequest only has the `NamespacedName`. How do you get the full object? Is this object cached, or result in an API call to the k8s master?API call

1. What metrics does the controller runtime emit? Describe what some of those metrics represent

1. What is leader election, and when would you use it? It is used to designate one instance of a controller as the leader to perform reconciliation, change cluster state, etc... It is used in distributed systems to prevent conflicts and ensure consistency

1. What is the difference between Kubebuilder and controller runtime?