#ifndef LEAKY_H
#define LEAKY_H

typedef struct {
    char* name;
    int id;
} User;

User* createUser(const char* name, int id);
void freeUser(User* user);
void freeUserCorrectly(User* user);
void printUser(User* user);

#endif