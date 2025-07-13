// Decoded bootloader.rom with comments.

c000 JMP 0xc003
c003 LXI SP 0x8ee0 // Initialize stack.
c006 MVI A, 0x82   // Confiugure IO controller.
c008 STA 0xff03
c00b JMP 0xc444    // Call the main func.
c00e NOP
c00f NOP

// func: fill display buffer with the value in 0x8ffa
c010 PUSH HL
c011 PUSH BC
c012 LXI HL 0x0000 // Store current stack pointer.
c015 DAD SP
c016 SHLD 0x8ff6
c019 LXI SP 0xc000 // Reset stack pointer to the end of the display buffer.
c01c LHLD 0x8ffa
c01f LXI BC 0x0300 // Fill the display buffer.
c022 PUSH HL
c023 PUSH HL
c024 PUSH HL
c025 PUSH HL
c026 PUSH HL
c027 PUSH HL
c028 PUSH HL
c029 PUSH HL
c02a DCX BC
c02b MOV A, B
c02c ORA C
c02d JCnd Z 0xc022
c030 LHLD 0x8ff6  // Restore the stack pointer.
c033 SPHL
c034 POP BC
c035 POP HL
c036 RET

// func: ??
c037 PUSH HL
c038 PUSH DE
c039 PUSH BC
c03a PUSH SP
c03b MOV A, C
c03c CPI 0x21
c03e JCnd c 0xc0d4
c041 LHLD 0x8ffc
c044 MOV A, H
c045 CPI 0xbe
c047 JCnd C 0xc0b2
c04a ADI 0x03
c04c STA 0x8ffd
c04f XCHG
c050 MOV A, C
c051 STA 0x8fe9
c054 SUI 0x20
c056 LHLD 0x8fe7
c059 ADD L
c05a MOV L, A
c05b DAD HL
c05c DAD HL
c05d DAD HL
c05e XCHG
c05f NOP
c060 MOV A, H
c061 ANI 0x03
c063 MOV C, A
c064 MVI A, 0x05
c066 SUB C
c067 MOV C, A
c068 MOV A, H
c069 ANI 0xfc
c06b RAR
c06c RAR
c06d ADI 0x90
c06f MOV H, A
c070 SHLD 0x8ff8
c073 MVI B, 0x08
c075 NOP
c076 LDAX DE
c077 MOV L, A
c078 MVI H, 0x00
c07a MOV A, C
c07b DAD HL
c07c DAD HL
c07d DCR A
c07e JCnd Z 0xc07b
c081 PUSH HL
c082 INX DE
c083 DCR B
c084 JCnd Z 0xc076
c087 MVI B, 0x08
c089 LHLD 0x8ff8
c08c POP DE
c08d MOV A, D
c08e CALL 0xc163
c091 MOV M, A
c092 INR H
c093 MOV A, E
c094 CALL 0xc163
c097 MOV M, A
c098 DCR H
c099 DCR L
c09a DCR B
c09b JCnd Z 0xc08c
c09e POP SP
c09f POP BC
c0a0 POP DE
c0a1 POP HL
c0a2 RET
c0a3 MOV C, A
c0a4 LDA 0x8ffa
c0a7 ORA A
c0a8 JCnd Z 0xc0af
c0ab MOV A, C
c0ac CMA
c0ad ANA M
c0ae RET
c0af MOV A, M
c0b0 ORA C
c0b1 RET
c0b2 MOV A, L
c0b3 CPI 0xf5
c0b5 JCnd C 0xc0c4
c0b8 ADI 0x0a
c0ba MOV L, A
c0bb STA 0x8ffc
c0be MVI H, 0x00
c0c0 MOV A, H
c0c1 JMP 0xc04a
c0c4 CALL 0xc22d
c0c7 NOP
c0c8 CALL 0xc010
c0cb LXI HL 0x0008
c0ce SHLD 0x8ffc
c0d1 JMP 0xc041
c0d4 LHLD 0x8ffc
c0d7 CPI 0x20
c0d9 JCnd z 0xc107
c0dc CPI 0x0a
c0de JCnd z 0xc113
c0e1 CPI 0x0d
c0e3 JCnd z 0xc128
c0e6 CPI 0x18
c0e8 JCnd z 0xc107
c0eb CPI 0x08
c0ed JCnd z 0xc12d
c0f0 CPI 0x19
c0f2 JCnd z 0xc139
c0f5 CPI 0x1a
c0f7 JCnd z 0xc145
c0fa CPI 0x0c
c0fc JCnd z 0xc151
c0ff CPI 0x1f
c101 JCnd z 0xc157
c104 JMP 0xc15d
c107 MOV A, H
c108 CPI 0xbe
c10a JCnd C 0xc113
c10d ADI 0x03
c10f MOV H, A
c110 JMP 0xc15d
c113 MVI H, 0x00
c115 MOV A, L
c116 CPI 0xf5
c118 JCnd C 0xc121
c11b ADI 0x0a
c11d MOV L, A
c11e JMP 0xc15d
c121 CALL 0xc22d
c124 NOP
c125 JMP 0xc157
c128 MVI H, 0x00
c12a JMP 0xc15d
c12d MOV A, H
c12e CPI 0x02
c130 JCnd c 0xc15d
c133 SUI 0x03
c135 MOV H, A
c136 JMP 0xc15d
c139 MOV A, L
c13a CPI 0x11
c13c JCnd c 0xc15d
c13f SUI 0x0a
c141 MOV L, A
c142 JMP 0xc15d
c145 MOV A, L
c146 CPI 0xf5
c148 JCnd C 0xc15d
c14b ADI 0x0a
c14d MOV L, A
c14e JMP 0xc15d
c151 LXI HL 0x0008
c154 JMP 0xc15d
c157 CALL 0xc010
c15a JMP 0xc151
c15d SHLD 0x8ffc
c160 JMP 0xc09e
c163 MOV C, A
c164 LDA 0x8fe9
c167 CPI 0x7f
c169 MOV A, C
c16a JCnd z 0xc0a3
c16d XRA M
c16e RET
c16f NOP

// func: ???
c170 PUSH HL
c171 PUSH BC
c172 LHLD 0x8ff1
c175 MVI A, 0x0b // io inputs: B,CHigh,CLow
c177 STA 0xff03
c17a CALL 0xc18f // wait
c17d MVI A, 0x0a // io inputs: B,CHigh
c17f STA 0xff03
c182 CALL 0xc18f // wait
c185 DCR H
c186 JCnd Z 0xc175
c189 NOP
c18a NOP
c18b NOP
c18c POP BC
c18d POP HL
c18e RET

// func: loop L times
c18f MOV B, L
c190 DCR B
c191 JCnd Z 0xc190
c194 RET

c195 PUSH SP
c196 MVI A, 0x40  // interact with port B
c198 STA 0x8ff1
c19b CALL 0xc170
c19e POP SP
c19f RET

c1a0 PUSH HL
c1a1 MVI A, 0x40  // interact with port B
c1a3 STA 0x8ff1
c1a6 CALL 0xc170
c1a9 POP HL
c1aa RET

// func: ???
c1ab PUSH HL
c1ac MVI A, 0x50  // interact with port B
c1ae JMP 0xc1a3

// func: ???
c1b1 PUSH HL
c1b2 PUSH BC
c1b3 MVI B, 0xff
c1b5 CALL 0xc283
c1b8 LDA 0x8ff4
c1bb ORA A
c1bc JCnd z 0xc1e2
c1bf CALL 0xc254
c1c2 LDA 0xff01
c1c5 ANI 0x02
c1c7 JCnd z 0xc24b
c1ca CALL 0xc25a
c1cd LDA 0xff00
c1d0 CPI 0xff
c1d2 JCnd Z 0xc1ff
c1d5 LDA 0xff02
c1d8 ORI 0xf0
c1da CPI 0xff
c1dc JCnd Z 0xc205
c1df JMP 0xc1b5
c1e2 CALL 0xc254
c1e5 LDA 0xff01
c1e8 ANI 0x02
c1ea JCnd Z 0xc1f2
c1ed MVI B, 0xff
c1ef JMP 0xc1ca
c1f2 DCR B
c1f3 JCnd Z 0xc1ca
c1f6 STA 0x8ff4
c1f9 CALL 0xc1a0
c1fc JMP 0xc1b5
c1ff MOV L, A
c200 MVI H, 0xff
c202 JMP 0xc208
c205 MOV H, A
c206 MVI L, 0xff
c208 MVI C, 0xfb
c20a INR C
c20b DAD HL
c20c JCnd c 0xc20a
c20f MOV L, C
c210 MVI B, 0xff
c212 CALL 0xc254
c215 LDA 0xff01
c218 ORI 0x03
c21a CPI 0xff
c21c JCnd Z 0xc235
c21f DCR B
c220 JCnd Z 0xc212
c223 JMP 0xc1b5
c226 LDA 0xff01
c229 CMA
c22a ANI 0xfe
c22c RET

// func: ???
c22d CALL 0xc226
c230 Ccnd Z 0xc260
c233 RET
c234 NOP

// func: load A with a value from ROM that corresponds to a particular bit...
c235 MVI C, 0xfd
c237 INR C
c238 RAR
c239 JCnd c 0xc237
c23c MOV A, C
c23d RAL
c23e RAL
c23f RAL
c240 RAL
c241 ADI 0xa0
c243 ORA L
c244 MOV L, A
c245 MVI H, 0xc4
c247 MOV A, M
c248 POP BC
c249 POP HL
c24a RET

// func: ???
c24b STA 0x8ff4
c24e CALL 0xc1ab
c251 JMP 0xc1b5

// func: io inputs=B
c254 MVI A, 0x82
c256 STA 0xff03
c259 RET

// func: io inputs=A,CLow
c25a MVI A, 0x91
c25c STA 0xff03
c25f RET

// func: ???
c260 PUSH BC
c261 LDA 0x8fef
c264 CPI 0x80
c266 JCnd z 0xc346
c269 MVI C, 0xff
c26b CALL 0xc254
c26e LDA 0xff01
c271 ORI 0x03
c273 CPI 0xff
c275 JCnd Z 0xc269
c278 MVI B, 0x15  // wait
c27a CALL 0xc190
c27d DCR C
c27e JCnd Z 0xc26b
c281 POP BC
c282 RET

// func: ???
c283 LHLD 0x8fed
c286 PCHL
c287 PUSH HL
c288 LXI HL 0x8feb
c28b DCR M
c28c Ccnd z 0xc291
c28f POP HL
c290 RET

c291 PUSH BC
c292 LXI HL 0x8feb
c295 MVI M, 0xff
c297 DCX HL
c298 INR M
c299 LHLD 0x8ffc
c29c INX HL
c29d INX HL
c29e SHLD 0x8ffc
c2a1 MVI C, 0x5f
c2a3 CALL 0xc2bc
c2a6 LHLD 0x8ffc
c2a9 DCX HL
c2aa DCX HL
c2ab SHLD 0x8ffc
c2ae POP BC
c2af RET
c2b0 PUSH HL
c2b1 PUSH SP
c2b2 LDA 0x8fea
c2b5 RAR
c2b6 Ccnd C 0xc291
c2b9 POP SP
c2ba POP HL
c2bb RET
c2bc PUSH HL
c2bd PUSH DE
c2be PUSH BC
c2bf PUSH SP
c2c0 MOV A, C
c2c1 LHLD 0x8ffc
c2c4 XCHG
c2c5 JMP 0xc050
c2c8 JMP 0xc35e
c2cb MVI A, 0x00
c2cd STA 0x8fea
c2d0 CALL 0xc1b1
c2d3 CALL 0xc2b0
c2d6 CPI 0x80
c2d8 JCnd z 0xc30c
c2db CALL 0xc195
c2de CPI 0x81
c2e0 JCnd C 0xc315
c2e3 STA 0x8ff0
c2e6 STA 0x8fef
c2e9 PUSH SP
c2ea CPI 0x21
c2ec JCnd c 0xc2fb
c2ef LDA 0x8ff4
c2f2 ORA A
c2f3 JCnd z 0xc300
c2f6 CPI 0x04
c2f8 JCnd z 0xc300
c2fb POP SP
c2fc RET
c2fd NOP
c2fe NOP
c2ff NOP
c300 POP SP
c301 CPI 0x40
c303 JCnd c 0xc309
c306 XRI 0x20
c308 RET
c309 XRI 0x10
c30b RET
c30c STA 0x8fef
c30f LDA 0x8ff0
c312 JMP 0xc2e9
c315 CPI 0x81
c317 JCnd Z 0xc325
c31a MVI A, 0x04
c31c STA 0x8ff4
c31f CALL 0xc1ab
c322 JMP 0xc2c8
c325 CPI 0x8c
c327 JCnd Z 0xc36c
c32a LXI HL 0xffff
c32d SHLD 0x8ffa
c330 JMP 0xc2c8
c333 LHLD 0x8fe5
c336 PCHL

// func: ??
c337 PUSH BC
c338 PUSH DE
c339 PUSH HL
c33a LXI HL 0x8feb
c33d MVI M, 0x01
c33f CALL 0xc2c8
c342 POP HL
c343 POP DE
c344 POP BC
c345 RET

c346 MVI C, 0x10
c348 MVI B, 0xff
c34a CALL 0xc190
c34d DCR C
c34e JCnd Z 0xc348
c351 JMP 0xc281
c354 PUSH BC
c355 MVI B, 0x15
c357 CALL 0xc190
c35a POP BC
c35b JMP 0xc287
c35e LXI HL 0x8feb
c361 MVI M, 0x01
c363 CALL 0xc283
c366 CALL 0xc260
c369 JMP 0xc2cb
c36c CPI 0x8b
c36e JCnd Z 0xc333
c371 LXI HL 0x0000
c374 JMP 0xc32d
c377 PUSH BC
c378 PUSH DE
c379 MVI C, 0x00 // <-- from the main func loop.
c37b MOV D, A
c37c LDA 0xff01  // Read tape reader data (first pin of port B).
c37f ANI 0x01
c381 MOV E, A    // Store it in E.
c382 MOV A, C
c383 ANI 0x7f
c385 RAL
c386 MOV C, A    // unsigned shift left of C
c387 LDA 0xff01  // Get port B: keyboard rows and tape reader.
c38a CPI 0x80
c38c JCnd c 0xc45a // Jump if the first row is pressed?
c38f ANI 0x01
c391 CMP E          // Check if there is a change on the tape reader.
c392 JCnd z 0xc387  // Jump if there is no change.
c395 ORA C          // Keep new bit to C which was shifted before.
c396 MOV C, A
c397 CALL 0xc3c9    // wait for 0x8fff
c39a LDA 0xff01     // Get the bit from the tape reader again, store in E.
c39d ANI 0x01
c39f MOV E, A
c3a0 MOV A, D
c3a1 ORA A
c3a2 JCnd S 0xc3be  // Jump if D starts with 0.
c3a5 MOV A, C
c3a6 CPI 0xe6
c3a8 JCnd Z 0xc3b2
c3ab XRA A
c3ac STA 0x8ff8
c3af JMP 0xc3bc
c3b2 CPI 0x19
c3b4 JCnd Z 0xc382
c3b7 MVI A, 0xff
c3b9 STA 0x8ff3
c3bc MVI D, 0x09
c3be DCR D
c3bf JCnd Z 0xc382
c3c2 LDA 0x8ff3
c3c5 XRA C
c3c6 POP DE
c3c7 POP BC
c3c8 RET

// func: wait for 0x8fff
c3c9 LDA 0x8fff
c3cc MOV B, A
c3cd JMP 0xc190

c3d0 PUSH BC
c3d1 PUSH DE
c3d2 PUSH SP
c3d3 MOV D, A
c3d4 MVI C, 0x08
c3d6 MOV A, D
c3d7 RAL
c3d8 MOV D, A
c3d9 ANI 0x01
c3db ORI 0x0e
c3dd STA 0xff03
c3e0 MOV E, A
c3e1 CALL 0xc484
c3e4 MOV A, E
c3e5 XRI 0x01
c3e7 STA 0xff03
c3ea CALL 0xc484
c3ed DCR C
c3ee JCnd Z 0xc3d6
c3f1 POP SP
c3f2 POP DE
c3f3 POP BC
c3f4 RET
c3f5 NOP
c3f6 NOP
c3f7 NOP
c3f8 NOP

// func: ???
c3f9 MVI A, 0xff
c3fb CALL 0xc377
c3fe MOV L, A
c3ff MVI A, 0x08
c401 CALL 0xc377
c404 MOV H, A
c405 SHLD 0x8fe3
c408 MVI A, 0x08
c40a CALL 0xc377
c40d MOV E, A
c40e MVI A, 0x08
c410 CALL 0xc377
c413 MOV D, A
c414 MVI A, 0x08
c416 CALL 0xc377
c419 MOV M, A
c41a CALL 0xc427
c41d INX HL
c41e JCnd Z 0xc414
c421 RET
c422 MVI A, 0xff
c424 JMP 0xc416

// func: part of mem copy below, check if we've reached DE.
c427 MOV A, H
c428 CMP D
c429 Rcnd Z
c42a MOV A, L
c42b CMP E
c42c RET

// func: mem copy from HL:DE to BC.
c42d MOV A, M
c42e STAX BC
c42f INX HL
c430 INX BC
c431 CALL 0xc427
c434 JCnd Z 0xc42d
c437 RET

// func
c438 MOV A, M   // Compare current HL with 0...
c439 MOV C, A
c43a CPI 0x00
c43c Rcnd z
c43d CALL 0xc037
c440 INX HL
c441 JMP 0xc438

// Subroutine called at the start.
c444 LXI HL 0xc473
c447 LXI DE 0xc490
c44a LXI BC 0x8fe3
c44d CALL 0xc42d   // Copy data to the user RAM. First bytes are address of ROM.
c450 CALL 0xc438
c453 CALL 0xc3f9
c456 LHLD 0x8fe3  // Original bytes at the ROM address.
c459 PCHL // Must be a loop to the beginning.

// func: ???
c45a CALL 0xc337
c45d CPI 0x0d
c45f JCnd z 0xc800 // Jump to the monitor.
c462 CPI 0x0a
c464 JCnd z 0xc46f // Jump to a program referred in 0x8fe1
c467 STA 0x8fff    // Current A will defined the next wait time.
c46a MVI A, 0xff
c46c JMP 0xc379    // Jump to reading the inputs.

c46f LHLD 0x8fe1
c472 PCHL

// Data region.
c473 NOP // c000, c2c8, ....
c474 Rcnd Z
c475 Rcnd z
c476 JCnd Z 0x18a0
c479 NOP
c47a NOP
c47b NOP
c47c NOP
c47d MOV D, H
c47e JMP 0x2020
c481 MOV B, B
c482 MOV B, B
c483 NOP
c484 LDA 0x8ffe
c487 JMP 0xc3cc
c48a NOP
c48b NOP
c48c NOP
c48d NOP