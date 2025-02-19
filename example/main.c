#include <stdio.h>
#include <stdlib.h>

void swap(int *a, int *b) {
    int tmp = *a;
    *a = *b;
    *b = tmp;
}

void quicksort(int *data, int l, int r) {
    if (l > r) return;

    int i = l;
    int j = r;
    int pivot = data[(l + r) / 2];

    while (i <= j) {
        while (data[i] < pivot) i++;
        while (data[j] > pivot) j--;

        if (i <= j) {
            swap(&data[i], &data[j]);

            i++;
            j--;
        }
    }

    quicksort(data, l, j);
    quicksort(data, i, r);
}

int main() {
    int n;
    scanf("%d", &n);

    int* data = (int*) malloc((n) * sizeof(int));
    if (data == NULL) {
        return 1;
    }

    for (int i = 0; i < n; i++) {
        scanf("%d", &data[i]);
    }

    quicksort(data, 0, n-1);

    for (int i = 0; i < n; i++) {
        printf("%d ", data[i]);
    }
}
