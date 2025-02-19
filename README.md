# Example

We have code and input for example task sorting in `./example`.

### Here is request structure with input:

```bash
curl -X POST -d '{"code": "#include <stdio.h>\n#include <stdlib.h>\n\nvoid swap(int *a, int *b) {\n    int tmp = *a;\n    *a = *b;\n    *b = tmp;\n}\n\nvoid quicksort(int *data, int l, int r) {\n    if (l > r) return;\n\n    int i = l;\n    int j = r;\n    int pivot = data[(l + r) / 2];\n\n    while (i <= j) {\n        while (data[i] < pivot) i++;\n        while (data[j] > pivot) j--;\n\n        if (i <= j) {\n            swap(&data[i], &data[j]);\n\n            i++;\n            j--;\n        }\n    }\n\n    quicksort(data, l, j);\n    quicksort(data, i, r);\n}\n\nint main() {\n    int n;\n    scanf(\"%d\", &n);\n\n    int* data = (int*) malloc((n) * sizeof(int));\n    if (data == NULL) {\n        return 1;\n    }\n\n    for (int i = 0; i < n; i++) {\n        scanf(\"%d\", &data[i]);\n    }\n\n    quicksort(data, 0, n-1);\n\n    for (int i = 0; i < n; i++) {\n        printf(\"%d \", data[i]);\n    }\n}","input":"5\n7 4 0 8 -12"}' localhost:2111/run
```

This request sends source file with exact content as in `./example/main.c` and input from `./example/input.txt`.


Feedback from server is in following format:

```json
{
    "status": "int",
    "stdout": "string",
    "stderr": "string"
}
```
