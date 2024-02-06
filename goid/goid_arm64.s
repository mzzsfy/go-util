#include "textflag.h"

TEXT ·getG(SB), NOSPLIT, $0
    MOVD    g, ret+0(FP)
    RET
