package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var kClientset *kubernetes.Clientset

// https://stackoverflow.com/a/51270134
func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func setup() {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error %s", err)
	}
	kClientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error %s", err)
	}

}

func verifyToken(clientId string) (bool, error) {
	ctx := context.TODO()
	tokenReview := authv1.TokenReview{
		Spec: authv1.TokenReviewSpec{
			Token:     clientId,
			Audiences: []string{"service-2"},
		},
	}
	result, err := kClientset.AuthenticationV1().TokenReviews().Create(ctx, &tokenReview, metav1.CreateOptions{})
	if err != nil {
		return false, err
	}
	log.Printf("%s\n", prettyPrint(result.Status))

	if result.Status.Authenticated {
		return true, nil
	}
	return false, nil
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	clientId := r.Header.Get("X-Client-Id")
	if len(clientId) == 0 {
		http.Error(w, "X-Client-Id header is not present", http.StatusUnauthorized)
		return
	}
	authenticated, err := verifyToken(clientId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !authenticated {
		http.Error(w, "Invalid token", http.StatusForbidden)
		return
	}

	w.Write([]byte("Hello from service2. You have been authenticated!"))
}

func main() {

	setup()

	http.HandleFunc("/", handleIndex)
	port := os.Getenv("PORT")
	if len(port) == 0 {
		log.Fatal("PORT must be specified!")
	}
	http.ListenAndServe(port, nil)
}
