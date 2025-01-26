#include <stdio.h>

int main(void) {
    int n, a, b;
    scanf("%d", &n);

    for (size_t i = 0; i < n; ++i) {
        scanf("%d %d", &a, &b);
        printf("%d\n", a + b);
    }

    return 0;
}
