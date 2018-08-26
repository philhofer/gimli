#include "textflag.h"

// ROUND(x, y, z, tmp0, tmp1, tmp2) yields:
// x = x
// y = tmp1
// z = tmp0
#define ROUND(x, y, z, tmp0, tmp1, tmp2) \
	VPSLLD  $1, z, tmp0   	 \ // tmp0 = z<<1
	VPSHUFB X7, x, x     	 \ // x = rotl(x, 24)
	VPSLLD  $9,  y, tmp1  	 \
	VPSRLD  $23, y, y     	 \
	VPXOR   tmp0, x, tmp0 	 \ // tmp0 = x ^ (z<<1)
	VPOR    y, tmp1, y    	 \ // y = rotl(y, 9)
	VPAND   y, z, tmp1    	 \ // tmp1 = y&z
	VPSLLD  $2, tmp1, tmp1	 \ // tmp1 = (y&z)<<2
	VPAND   x, y, tmp2       \ // tmp2 = (x&y)
	VPXOR   tmp1, tmp0, tmp0 \ // tmp0 = newz = x ^ (z << 1) ^ ((y&z) << 2)
	VPOR    x, z, tmp1       \ // tmp1 = (x|z)
	VPSLLD  $1, tmp1, tmp1   \ // tmp1 = (x|z)<<1
	VPSLLD  $3, tmp2, tmp2   \ // tmp2 = (x&y)<<3
	VPXOR   x,  tmp1, tmp1   \ // ... = x ^ ((x|z)<<1)
	VPXOR   y,  tmp1, tmp1   \ // tmp1 = newy = y ^ x ^ ((x|z)<<1)
	VPXOR   tmp2, y, tmp2    \ // tmp2 = y ^ (x&y)<<3
	VPXOR   tmp2, z, x         // x = z ^ y ^ ((x&y)<<3)

// for doing a 32x4 rotl(x, 24),
// we use VPSHUFB and a precomputed table
DATA shuftab<>+0(SB)/1, $1
DATA shuftab<>+1(SB)/1, $2
DATA shuftab<>+2(SB)/1, $3
DATA shuftab<>+3(SB)/1, $0
DATA shuftab<>+4(SB)/1, $5
DATA shuftab<>+5(SB)/1, $6
DATA shuftab<>+6(SB)/1, $7
DATA shuftab<>+7(SB)/1, $4
DATA shuftab<>+8(SB)/1, $9
DATA shuftab<>+9(SB)/1, $10
DATA shuftab<>+10(SB)/1, $11
DATA shuftab<>+11(SB)/1, $8
DATA shuftab<>+12(SB)/1, $13
DATA shuftab<>+13(SB)/1, $14
DATA shuftab<>+14(SB)/1, $15
DATA shuftab<>+15(SB)/1, $12
GLOBL shuftab<>(SB), RODATA|NOPTR, $16

DATA coeffs<>+00(SB)/4, $0x9e377904
DATA coeffs<>+16(SB)/4, $0x9e377908
DATA coeffs<>+32(SB)/4, $0x9e37790c
DATA coeffs<>+48(SB)/4, $0x9e377910
DATA coeffs<>+64(SB)/4, $0x9e377914
DATA coeffs<>+80(SB)/4, $0x9e377918
GLOBL coeffs<>(SB), RODATA|NOPTR, $96

// func roundAVX(state *[12]uint32)
TEXT ·roundAVX(SB),NOSPLIT,$0
	MOVQ  state+0(FP), SI
	MOVUPS 00(SI), X0
	MOVUPS 16(SI), X1
	MOVUPS 32(SI), X2
	CALL   gimliregs<>(SB)
	MOVUPS X0, 00(SI)
	MOVUPS X1, 16(SI)
	MOVUPS X2, 32(SI)
	RET

// func hashroundsAVX(state *[12]uin32, src []byte, rounds int)
TEXT ·hashroundsAVX(SB),NOSPLIT,$0
	MOVQ state+0(FP), SI
	MOVQ src+8(FP), DI
	MOVQ rounds+32(FP), R9
	MOVUPS 00(SI), X0
	MOVUPS 16(SI), X1
	MOVUPS 32(SI), X2
	JMP     test
loop:
	SUBQ  $1, R9
	VPXOR 0(DI), X0, X0
	CALL  gimliregs<>(SB)
	ADDQ  $16, DI
test:
	CMPQ R9, $0
	JNE  loop
	MOVUPS X0, 00(SI)
	MOVUPS X1, 16(SI)
	MOVUPS X2, 32(SI)
	RET

// in/out: x=X0, y=X1, z=X2
// clobbers: AX, DX, X*
TEXT gimliregs<>(SB),NOSPLIT,$0
	LEAQ   coeffs<>+0(SB), AX
	LEAQ   -96(AX), DX
	MOVUPS shuftab<>+0(SB), X7
loop:
	SUBQ $16, AX
	ROUND(X0, X1, X2, X3, X4, X5)
	VPSHUFD $177, X0, X0
	VPXOR   96(AX), X0, X0
	ROUND(X0, X4, X3, X2, X1, X5)
	ROUND(X0, X1, X2, X3, X4, X5)
	VPSHUFD $78, X0, X0
	ROUND(X0, X4, X3, X2, X1, X5)
	CMPQ AX, DX
	JNE  loop
	RET
