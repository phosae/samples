#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include "leaky.h"

// This function allocates a User struct and initializes it
User* createUser(const char* name, int id) {
    User* user = (User*)malloc(sizeof(User));
    if (user == NULL) {
        return NULL; // Handle allocation failure
    }

    user->name = (char*)malloc(strlen(name) + 1);
    if (user->name == NULL) {
        free(user);
        return NULL; // Handle allocation failure
    }
    strcpy(user->name, name);
    user->id = id;

    printf("C: Created user: %s, id: %d at %p\n", user->name, user->id, user);
    fflush(stdout);
    return user;
}

// This function is INTENDED to free the User struct, but it has a bug.
void freeUser(User* user) {
    printf("C: Freeing user at %p\n", user);
    fflush(stdout);
    // BUG: We're only freeing the name, not the User struct itself!
    free(user->name);
    // We should be freeing the user struct as well: free(user);
}

// Correctly frees the User struct
void freeUserCorrectly(User* user) {
    printf("C: Freeing user at %p\n", user);
    fflush(stdout);
    free(user->name);
    free(user); // Now we also free the User struct itself
}

// This is just for demo purposes; imagine a more complex C function here.
void printUser(User* user) {
    if (user != NULL) {
      printf("C: User: %s, id: %d\n", user->name, user->id);
      fflush(stdout);
    }
}