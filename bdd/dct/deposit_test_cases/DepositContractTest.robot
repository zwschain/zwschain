*** Settings ***
Resource          publicParams.txt
Library           RequestsLibrary
Library           String

*** Variables ***
${mediatorAddr_01}    ${EMPTY}
${foundationAddr}    ${EMPTY}
${mediatorAddr_02}    ${EMPTY}
${juryAddr_01}    ${EMPTY}
${developerAddr_01}    ${EMPTY}
${juryAddr_02}    ${EMPTY}
${developerAddr_02}    ${EMPTY}
${juryAddr_01_pubkey}    ${EMPTY}
${juryAddr_02_pubkey}    ${EMPTY}
${m1_pubkey}      ${EMPTY}
${m2_pubkey}      ${EMPTY}
${votedAddress}    ${EMPTY}
${votedAddress01}    ${EMPTY}
${votedAddress02}    ${EMPTY}
${votedAddress03}    ${EMPTY}
${votedAddress04}    ${EMPTY}
${votedAddress05}    ${EMPTY}
${votedAddress06}    ${EMPTY}

*** Test Cases ***
Business_01
    [Documentation]    mediator 交付 50 ptn 才可以加入候选列表
    ...
    ...    mediator节点申请并且进入申请列表-》基金会同意并移除申请列表-》进入同意列表-》节点加入保证金（不够足够）无法进入候选列表-》节点交够保证金-》进入候选列表（自动进入Jury候选列表）-》节点申请退出并且进入退出列表-》基金会同意并移除节点候选列表（Jury候选列表）和退出列表。结果：同意列表包含该地址
    ${amount}    getBalance    ${mediatorAddr_01}    PTN
    log    ${amount}    #999949995
    #    Should Be Equal As Numbers
    Should Be Equal As Numbers    ${amount}    10000
    ${amount}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM    PTN
    log    ${amount}    #0
    Should Be Equal As Numbers    ${amount}    0
    ${result}    applyBecomeMediator    ${mediatorAddr_01}    ${m1_pubkey}    #1
    log    ${result}
    ${addressMap1}    getBecomeMediatorApplyList
    log    ${addressMap1}
    Dictionary Should Contain Key    ${addressMap1}    ${mediatorAddr_01}    #有该节点
    ${result}    handleForApplyBecomeMediator    ${foundationAddr}    ${mediatorAddr_01}    Ok    #基金会处理列表里的节点（同意）    #1
    log    ${result}
    ${addressMap2}    getAgreeForBecomeMediatorList
    log    ${addressMap2}
    Dictionary Should Contain Key    ${addressMap2}    ${mediatorAddr_01}    #有该节点
    ${result}    mediatorPayToDepositContract    ${mediatorAddr_01}    30    #在同意列表里的节点，可以交付保证金    #这里交的数量不是规定的保证金数量，导致无法加入候选列表，并且相应保证金退还该地址    #1
    log    ${result}
    ${amount}    getBalance    ${mediatorAddr_01}    PTN    #499999992
    log    ${amount}
    Should Be Equal As Numbers    ${amount}    9998
    ${amount}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM    PTN
    log    ${amount}    #0
    Should Be Equal As Numbers    ${amount}    0
    ${addressMap3}    getListForMediatorCandidate
    log    ${addressMap3}
    Dictionary Should Not Contain Key    ${addressMap3}    ${mediatorAddr_01}    #无该节点
    ${result}    mediatorPayToDepositContract    ${mediatorAddr_01}    ${medDepositAmount}    #在同意列表里的节点，可以交付保证金    #51
    log    ${result}
    ${amount}    getBalance    ${mediatorAddr_01}    PTN
    log    ${amount}    #999949941
    Should Be Equal As Numbers    ${amount}    9947
    ${amount}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM    PTN
    log    ${amount}    #0
    Should Be Equal As Numbers    ${amount}    50
    ${addressMap3}    getListForMediatorCandidate
    log    ${addressMap3}
    Dictionary Should Contain Key    ${addressMap3}    ${mediatorAddr_01}    #有该节点
    ${resul}    getListForJuryCandidate    #mediator自动称为jury
    Dictionary Should Contain Key    ${resul}    ${mediatorAddr_01}    #有该节点
    log    ${resul}
    ${mDeposit}    getMediatorDepositWithAddr    ${mediatorAddr_01}    #获取该地址保证金账户详情
    log    ${mDeposit}
    Should Not Be Equal    ${mDeposit["balance"]}    0    #有余额
    GetAllMediators
    ${result}    applyQuitMediator    ${mediatorAddr_01}    MediatorApplyQuit    #该节点申请退出mediator候选列表    #1
    log    ${result}
    ${addressMap4}    getQuitMediatorApplyList
    log    ${addressMap4}
    Dictionary Should Contain Key    ${addressMap4}    ${mediatorAddr_01}    #有该节点
    ${result}    handleForApplyForQuitMediator    ${foundationAddr}    ${mediatorAddr_01}    Ok    HandleForApplyQuitMediator    #基金会处理退出候选列表里的节点（同意）
    ...    #1
    log    ${result}
    ${amount}    getBalance    ${mediatorAddr_01}    PTN    #99,999,970‬
    log    ${amount}
    Should Be Equal As Numbers    ${amount}    9996
    ${amount}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM    PTN
    log    ${amount}    #0
    Should Be Equal As Numbers    ${amount}    0
    ${resul}    getListForJuryCandidate    #mediator退出候选列表，则移除该jury
    Dictionary Should Not Contain Key    ${resul}    ${mediatorAddr_01}    #无该节点
    log    ${resul}
    ${mDeposit}    getMediatorDepositWithAddr    ${mediatorAddr_01}    #获取该地址保证金账户详情
    log    ${mDeposit}
    Should Be Equal    ${mDeposit["balance"]}    0    #账户地址存在
    ${result}    getBecomeMediatorApplyList
    log    ${result}
    Dictionary Should Not Contain Key    ${result}    ${mediatorAddr_01}    #无该节点
    ${result}    getAgreeForBecomeMediatorList
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${mediatorAddr_01}    #有该节点
    ${result}    getListForMediatorCandidate
    log    ${result}
    Dictionary Should Not Contain Key    ${result}    ${mediatorAddr_01}    #无该节点
    ${result}    getQuitMediatorApplyList
    log    ${result}
    Dictionary Should Not Contain Key    ${result}    ${mediatorAddr_01}    #无该节点
    GetAllMediators

Business_02
    [Documentation]    没收mediator节点
    ...
    ...    mediator节点申请并且进入申请列表-》基金会同意并移除申请列表-》进入同意列表-》节点交够保证金-》进入候选列表（自动进入Jury候选列表）-》某节点申请没收该mediator节点并进入没收列表-》基金会同意并移除候选列表，该节点的PTN转到基金会地址。结果：同意列表有该mediator节点地址，账户余额为0
    ${result}    applyBecomeMediator    ${mediatorAddr_02}    ${m2_pubkey}
    log    ${result}
    ${addressMap1}    getBecomeMediatorApplyList
    log    ${addressMap1}
    Dictionary Should Contain Key    ${addressMap1}    ${mediatorAddr_02}    #有该节点
    ${result}    handleForApplyBecomeMediator    ${foundationAddr}    ${mediatorAddr_02}    Ok    #基金会处理列表里的节点（同意）
    log    ${result}
    ${addressMap2}    getAgreeForBecomeMediatorList
    log    ${addressMap2}
    Dictionary Should Contain Key    ${addressMap2}    ${mediatorAddr_02}    #有该节点
    ${result}    mediatorPayToDepositContract    ${mediatorAddr_02}    ${medDepositAmount}    #在同意列表里的节点，可以交付保证金（大于或等于保证金数量）,需要200000000000及以上
    log    ${result}
    ${addressMap3}    getListForMediatorCandidate
    log    ${addressMap3}
    Dictionary Should Contain Key    ${addressMap3}    ${mediatorAddr_02}    #有该节点
    ${resul}    getListForJuryCandidate    #mediator自动称为jury
    Dictionary Should Contain Key    ${resul}    ${mediatorAddr_02}    #有该节点
    log    ${resul}
    ${mDeposit}    getMediatorDepositWithAddr    ${mediatorAddr_02}    #获取该地址保证金账户详情
    log    ${mDeposit}
    Should Not Be Equal    ${mDeposit["balance"]}    0    #有余额
    ${result}    applyForForfeitureDeposit    ${foundationAddr}    ${mediatorAddr_02}    Mediator    nothing to do    #某个地址申请没收该节点保证金（全部）
    log    ${result}
    ${result}    getListForForfeitureApplication
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${mediatorAddr_02}    #有该节点
    ${result}    handleForForfeitureApplication    ${foundationAddr}    ${mediatorAddr_02}    Ok    #基金会处理（同意），这是会移除mediator出候选列表
    log    ${result}
    ${result}    getMediatorDepositWithAddr    ${mediatorAddr_02}
    log    ${result}    #余额为 0
    Should Not Be Equal    ${result}    balance is nil    #不为空
    ${result}    getAgreeForBecomeMediatorList
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${mediatorAddr_02}    #同意列表有该地址
    ${result}    getListForMediatorCandidate
    log    ${result}
    Dictionary Should Not Contain Key    ${result}    ${mediatorAddr_02}    #候选列表无该地址
    ${result}    getListForForfeitureApplication
    log    ${result}
    Dictionary Should Not Contain Key    ${result}    ${mediatorAddr_02}    #没收列表无该地址
    ${resul}    getListForJuryCandidate    #mediator退出候选列表，则移除该jury
    Dictionary Should Not Contain Key    ${resul}    ${mediatorAddr_02}    #jury候选列表无该地址
    log    ${resul}
    GetAllMediators
    ${amount}    getBalance    ${juryAddr_01}    PTN
    log    ${amount}    #9998
    ${amount}    getBalance    ${juryAddr_02}    PTN
    log    ${amount}    #9989
    ${amount}    getBalance    ${developerAddr_01}    PTN
    log    ${amount}    #9998
    ${amount}    getBalance    ${developerAddr_02}    PTN
    log    ${amount}    #9998

Business_03
    [Documentation]    jury 交付 10 ptn 才可以加入候选列表
    ...
    ...    Jury 节点交付固定保证金并进入候选列表-》申请退出并进入退出列表-》基金会同意并移除候选列表，退出列表。
    ${resul}    juryPayToDepositContract    ${juryAddr_01}    10    ${juryAddr_01_pubkey}
    log    ${resul}
    ${result}    getJuryBalance    ${juryAddr_01}    #获取该地址保证金账户详情
    log    ${result}    #余额为100000000000000
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForJuryCandidate
    Dictionary Should Contain Key    ${resul}    ${juryAddr_01}    #候选列表有该地址
    log    ${resul}
    GetAllJury
    ${result}    applyQuitMediator    ${juryAddr_01}    JuryApplyQuit    #该节点申请退出mediator候选列表
    log    ${result}
    ${addressMap4}    getQuitMediatorApplyList    #获取申请mediator列表里的节点（不为空）
    log    ${addressMap4}
    Dictionary Should Contain Key    ${addressMap4}    ${juryAddr_01}
    ${result}    handleForApplyForQuitMediator    ${foundationAddr}    ${juryAddr_01}    Ok    HandleForApplyQuitJury    #基金会处理退出候选列表里的节点（同意）
    log    ${result}
    ${resul}    getListForJuryCandidate    #mediator退出候选列表，则移除该jury
    Dictionary Should Not Contain Key    ${resul}    ${juryAddr_01}
    log    ${resul}
    ${result}    getQuitMediatorApplyList    #为空
    log    ${result}
    Dictionary Should Not Contain Key    ${result}    ${juryAddr_01}
    ${result}    getJuryBalance    ${juryAddr_01}    #获取该地址保证金账户详情
    log    ${result}    #余额为100000000000000
    Should Be Equal    ${result}    balance is nil
    GetAllJury

Business_04
    [Documentation]    没收jury节点
    ...
    ...    Jury 节点交付固定保证金并进入候选列表-》某节点申请没收该jury节点并进入没收列表-》基金会同意并移除候选列表，退出列表，该节点的PTN转到基金会地址。
    ${resul}    juryPayToDepositContract    ${juryAddr_02}    10    ${juryAddr_02_pubkey}
    log    ${resul}
    ${result}    getJuryBalance    ${juryAddr_02}    #获取该地址保证金账户详情
    log    ${result}
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForJuryCandidate
    Dictionary Should Contain Key    ${resul}    ${juryAddr_02}    #候选列表有该地址
    log    ${resul}
    ${result}    applyForForfeitureDeposit    ${foundationAddr}    ${juryAddr_02}    Jury    nothing to do    #某个地址申请没收该节点保证金（全部）
    log    ${result}
    ${result}    getListForForfeitureApplication
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${juryAddr_02}    #没收列表有该地址
    ${result}    handleForForfeitureApplication    ${foundationAddr}    ${juryAddr_02}    Ok    #基金会处理（同意），这是会移除mediator出候选列表
    log    ${result}
    ${result}    getJuryBalance    ${juryAddr_02}
    log    ${result}
    Should Be Equal    ${result}    balance is nil    #不为空
    ${resul}    getListForJuryCandidate
    Dictionary Should Not Contain Key    ${resul}    ${juryAddr_02}    #候选列表无该地址
    log    ${resul}
    GetAllJury

Business_05
    [Documentation]    dev 交付 1 ptn 才可以加入合约开发者列表
    ...
    ...    Developer 节点交付固定保证金并进入列表-》申请退出并进入退出列表-》基金会同意并移除列表，退出列表。
    ${resul}    developerPayToDepositContract    ${developerAddr_01}    1
    log    ${resul}
    ${result}    getCandidateBalanceWithAddr    ${developerAddr_01}    #获取该地址保证金账户详情
    log    ${result}
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForDeveloperCandidate
    Dictionary Should Contain Key    ${resul}    ${developerAddr_01}    #候选列表无该地址
    log    ${resul}
    GetAllNodes
    ${result}    applyQuitMediator    ${developerAddr_01}    DeveloperApplyQuit    #该节点申请退出mediator候选列表
    log    ${result}
    ${addressMap4}    getQuitMediatorApplyList    #获取申请mediator列表里的节点（不为空）
    log    ${addressMap4}
    Dictionary Should Contain Key    ${addressMap4}    ${developerAddr_01}
    ${result}    handleForApplyForQuitMediator    ${foundationAddr}    ${developerAddr_01}    Ok    HandleForApplyQuitDev    #基金会处理退出候选列表里的节点（同意）
    log    ${result}
    ${resul}    getListForDeveloperCandidate    #mediator退出候选列表，则移除该jury
    Dictionary Should Not Contain Key    ${resul}    ${developerAddr_01}
    log    ${resul}
    ${result}    getQuitMediatorApplyList    #为空
    log    ${result}
    Dictionary Should Not Contain Key    ${result}    ${developerAddr_01}
    GetAllNodes

Business_06
    [Documentation]    没收dev节点
    ...
    ...    Developer节点交付固定保证金并进入候选列表-》某节点申请没收该Developer节点并进入没收列表-》基金会同意并移除候选列表，退出列表，该节点的PTN转到基金会地址。
    ${resul}    developerPayToDepositContract    ${developerAddr_02}    1
    log    ${resul}
    ${result}    getCandidateBalanceWithAddr    ${developerAddr_02}    #获取该地址保证金账户详情
    log    ${result}
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForDeveloperCandidate
    Dictionary Should Contain Key    ${resul}    ${developerAddr_02}    #候选列表无该地址
    log    ${resul}
    ${result}    applyForForfeitureDeposit    ${foundationAddr}    ${developerAddr_02}    Developer    nothing to do    #某个地址申请没收该节点保证金（全部）
    log    ${result}
    ${result}    getListForForfeitureApplication
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${developerAddr_02}    #没收列表有该地址
    ${result}    handleForForfeitureApplication    ${foundationAddr}    ${developerAddr_02}    Ok    #基金会处理（同意），这是会移除mediator出候选列表
    log    ${result}
    ${result}    getCandidateBalanceWithAddr    ${developerAddr_02}
    log    ${result}
    Should Be Equal    ${result}    balance is nil    #不为空
    ${resul}    getListForDeveloperCandidate
    Dictionary Should Not Contain Key    ${resul}    ${developerAddr_02}    #候选列表无该地址
    log    ${resul}
    GetAllNodes

Business_07
    [Documentation]    创建新Token，使用新Token交付保证金，由于保证金只支持PTN，所以交保证金失败
    ...
    ...    DPT+10E7XBWEBT0K2H2GE0O
    ...
    ...    DPT+10102JC6CQU8OK204BXA
    ${amount}    getBalance    ${mediatorAddr_02}    PTN
    log    ${amount}
    Should Be Equal As Numbers    ${amount}    9948
    ${amount}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM    PTN
    log    ${amount}    #0
    Should Be Equal As Numbers    ${amount}    0
    ${result}    createToken    ${mediatorAddr_02}    #1
    ${assetId}    ccquery
    log    ${assetId}
    ${amount}    getBalance    ${mediatorAddr_02}    ${assetId}
    log    ${amount}
    Should Be Equal As Numbers    ${amount}    1000
    ${result}    invokeToken    ${mediatorAddr_02}    ${assetId}    #在同意列表里的节点，可以交付保证金    #这里交的数量不是规定的保证金数量，导致无法加入候选列表，并且相应保证金退还该地址    #1
    log    ${result}
    ${addressMap3}    getListForMediatorCandidate
    log    ${addressMap3}
    Dictionary Should Not Contain Key    ${addressMap3}    ${mediatorAddr_02}    #无该节点
    ${amount}    getBalance    ${mediatorAddr_02}    ${assetId}
    log    ${amount}
    Should Be Equal As Numbers    ${amount}    1000
    ${amount}    getBalance    ${mediatorAddr_02}    PTN
    log    ${amount}
    Should Be Equal As Numbers    ${amount}    9946
    ${amount}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM    PTN
    log    ${amount}    #0
    Should Be Equal As Numbers    ${amount}    0

middle_cases
    [Documentation]    查询该阶段，保证金合约地址账户余额及各节点地址账户余额信息，及在保证金里面各节点的相关信息。
    log    Mediator
    ${addressMap1}    getBecomeMediatorApplyList
    log    ${addressMap1}    #空
    ${addressMap2}    getAgreeForBecomeMediatorList
    log    ${addressMap2}    #同意的2个
    ${addressMap3}    getListForMediatorCandidate
    log    ${addressMap3}    #默认的5个
    GetAllMediators    #加上默认的5个，一共7个
    log    Jury
    ${resul}    getListForJuryCandidate
    log    ${resul}    #默认的5个
    GetAllJury    #默认的5个
    log    Developer
    ${resul}    getListForDeveloperCandidate
    log    ${resul}    #空
    GetAllNodes    #空
    log    All
    ${addressMap4}    getQuitMediatorApplyList
    log    ${addressMap4}    #空
    ${list}    getListForForfeitureApplication
    log    ${list}    #空
    log    "Balance..."
    ${amount}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM    PTN
    log    ${amount}    #0
    ${amount}    getBalance    ${foundationAddr}    PTN
    log    ${amount}    #999930043
    ${amount}    getBalance    ${mediatorAddr_01}    PTN
    log    ${amount}    #9996
    ${amount}    getBalance    ${mediatorAddr_02}    PTN
    log    ${amount}    #9946
    ${amount}    getBalance    ${juryAddr_01}    PTN
    log    ${amount}    #9998，花了2个PTN手续费
    ${amount}    getBalance    ${juryAddr_02}    PTN
    log    ${amount}    #9989，少10个PTN是因为被没收的，花了1个PTN手续费
    ${amount}    getBalance    ${developerAddr_01}    PTN
    log    ${amount}    #9998，花了2个PTN手续费
    ${amount}    getBalance    ${developerAddr_02}    PTN
    log    ${amount}    #9998，被没收了1个PTN，花了1个PTN手续费

Business_09
    [Documentation]    jury 和 dev 交付保证金数量不对流程
    log    jury
    ${amount}    getBalance    ${juryAddr_01}    PTN
    log    ${amount}
    Should Be Equal As Numbers    ${amount}    9998
    sleep    1
    ${resul}    juryPayToDepositContract    ${juryAddr_01}    11    ${juryAddr_01_pubkey}    #应该是10，但是为11
    log    ${resul}
    sleep    5
    ${amount}    getBalance    ${juryAddr_01}    PTN
    log    ${amount}
    Should Be Equal As Numbers    ${amount}    9997
    log    dev
    sleep    1
    ${amount}    getBalance    ${developerAddr_01}    PTN
    log    ${amount}
    Should Be Equal As Numbers    ${amount}    9998
    sleep    1
    ${resul}    developerPayToDepositContract    ${developerAddr_01}    2    #应该为1，但是为2
    log    ${resul}
    sleep    5
    ${amount}    getBalance    ${developerAddr_01}    PTN
    log    ${amount}
    Should Be Equal As Numbers    ${amount}    9997
    sleep    1
    ${amount}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM    PTN
    log    ${amount}
    Should Be Equal As Numbers    ${amount}    0

Business_10
    [Documentation]    jury 交付 10 ptn 才可以加入候选列表
    ...
    ...    Jury 节点交付固定保证金并进入候选列表-》申请退出并进入退出列表-》基金不同意并移除候选列表，退出列表。
    ${resul}    juryPayToDepositContract    ${juryAddr_01}    10    ${juryAddr_01_pubkey}    #11
    log    ${resul}
    ${result}    getJuryBalance    ${juryAddr_01}    #获取该地址保证金账户详情
    log    ${result}
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForJuryCandidate
    Dictionary Should Contain Key    ${resul}    ${juryAddr_01}    #候选列表有该地址
    log    ${resul}
    ${result}    applyQuitMediator    ${juryAddr_01}    JuryApplyQuit    #该节点申请退出mediator候选列表    #1
    log    ${result}
    ${addressMap4}    getQuitMediatorApplyList    #获取申请mediator列表里的节点（不为空）
    log    ${addressMap4}
    Dictionary Should Contain Key    ${addressMap4}    ${juryAddr_01}
    ${result}    handleForApplyForQuitMediator    ${foundationAddr}    ${juryAddr_01}    no    HandleForApplyQuitJury    #基金会处理退出候选列表里的节点（同意）
    log    ${result}
    ${result}    getJuryBalance    ${juryAddr_01}    #获取该地址保证金账户详情
    log    ${result}
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForJuryCandidate    #mediator退出候选列表，则移除该jury
    Dictionary Should Contain Key    ${resul}    ${juryAddr_01}
    log    ${resul}
    ${result}    getQuitMediatorApplyList    #为空
    log    ${result}
    Dictionary Should Not Contain Key    ${result}    ${juryAddr_01}
    ${amount}    getBalance    ${juryAddr_01}    PTN
    log    ${amount}    #9985
    Should Be Equal As Numbers    ${amount}    9985
    ${amount}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM    PTN
    log    ${amount}    #108.66235，866,235,000是质押增发的，没有变化

Business_11
    [Documentation]    dev 交付 1 ptn 才可以加入合约开发者列表
    ...
    ...    Developer 节点交付固定保证金并进入列表-》申请退出并进入退出列表-》基金会不同意并移除列表，退出列表。
    ${resul}    developerPayToDepositContract    ${developerAddr_01}    1    #2
    log    ${resul}
    ${result}    getCandidateBalanceWithAddr    ${developerAddr_01}    #获取该地址保证金账户详情
    log    ${result}
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForDeveloperCandidate
    Dictionary Should Contain Key    ${resul}    ${developerAddr_01}    #候选列表无该地址
    log    ${resul}
    ${result}    applyQuitMediator    ${developerAddr_01}    DeveloperApplyQuit    #该节点申请退出mediator候选列表    #1
    log    ${result}
    ${addressMap4}    getQuitMediatorApplyList    #获取申请mediator列表里的节点（不为空）
    log    ${addressMap4}
    Dictionary Should Contain Key    ${addressMap4}    ${developerAddr_01}
    ${result}    handleForApplyForQuitMediator    ${foundationAddr}    ${developerAddr_01}    no    HandleForApplyQuitDev    #基金会处理退出候选列表里的节点（同意）
    log    ${result}
    ${resul}    getListForDeveloperCandidate    #mediator退出候选列表，则移除该jury
    Dictionary Should Contain Key    ${resul}    ${developerAddr_01}
    log    ${resul}
    ${result}    getQuitMediatorApplyList    #为空
    log    ${result}
    Dictionary Should Not Contain Key    ${result}    ${developerAddr_01}
    ${amount}    getBalance    ${developerAddr_01}    PTN
    log    ${amount}
    Should Be Equal As Numbers    ${amount}    9994
    ${result}    getCandidateBalanceWithAddr    ${developerAddr_01}    #获取该地址保证金账户详情
    log    ${result}
    Should Not Be Equal    ${result}    balance is nil
    ${amount}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM    PTN
    log    ${amount}    #108.66235，866,235,000是质押增发的，没有变化

Business_12
    [Documentation]    没收上一个测试用例的jury dev节点，但是基金会不同意
    log    jury
    ${result}    applyForForfeitureDeposit    ${foundationAddr}    ${juryAddr_01}    Jury    nothing to do    #某个地址申请没收该节点保证金（全部）
    log    ${result}
    ${result}    getListForForfeitureApplication
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${juryAddr_01}    #没收列表有该地址
    ${result}    handleForForfeitureApplication    ${foundationAddr}    ${juryAddr_01}    no    #基金会处理（同意），这是会移除mediator出候选列表
    log    ${result}
    ${result}    getJuryBalance    ${juryAddr_01}
    log    ${result}
    Should Not Be Equal    ${result}    balance is nil    #不为空
    ${resul}    getListForJuryCandidate
    Dictionary Should Contain Key    ${resul}    ${juryAddr_01}    #候选列表无该地址
    log    ${resul}
    ${amount}    getBalance    ${juryAddr_01}    PTN
    log    ${amount}
    Should Be Equal As Numbers    ${amount}    9985
    ${result}    getListForForfeitureApplication
    log    ${result}
    log    dev
    ${result}    applyForForfeitureDeposit    ${foundationAddr}    ${developerAddr_01}    Developer    nothing to do    #某个地址申请没收该节点保证金（全部）
    log    ${result}
    ${result}    getListForForfeitureApplication
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${developerAddr_01}    #没收列表有该地址
    ${result}    handleForForfeitureApplication    ${foundationAddr}    ${developerAddr_01}    no    #基金会处理（同意），这是会移除mediator出候选列表
    log    ${result}
    ${result}    getCandidateBalanceWithAddr    ${developerAddr_01}
    log    ${result}
    Should Not Be Equal    ${result}    balance is nil    #不为空
    ${resul}    getListForDeveloperCandidate
    Dictionary Should Contain Key    ${resul}    ${developerAddr_01}    #候选列表无该地址
    log    ${resul}
    ${amount}    getBalance    ${developerAddr_01}    PTN
    log    ${amount}
    Should Be Equal As Numbers    ${amount}    9994
    ${result}    getListForForfeitureApplication
    log    ${result}
    ${amount}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM    PTN
    log    ${amount}    #108.66235，866,235,000是质押增发的，没有变化

PledgeTest02
    [Documentation]    1.3个地址质押，每个地址质押100ptn
    ...    2.第一次将3个地址添加
    ...    3.继续有3个新地址质押，每个地址质押100ptn，这个时候，前面有两个地址分别继续质押并赎回相同ptn数量
    ...    4.这里第一次分红：质押总量=30000000000，日分红总量=288745000，得288745000/30000000000=0.0096248333333333
    ...    5.这里是第二次分红：质押总量=30000000000+30000000000+288745000=60288744999，日分红总量=288745000，得288745000/60288744999=0.0047893682312476
    ${depositOne}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM    PTN
    log    ${depositOne}
    ${result}    queryPledgeStatusByAddr    ${votedAddress01}    #查看某地址的质押结果
    log    ${result}
    ${result}    QueryPledgeHistoryByAddr    ${votedAddress01}
    log    ${result}
    ${result}    getBalance    ${votedAddress01}    PTN
    log    ${result}
    Should Be Equal As Numbers    ${result}    10000
    ${result}    queryPledgeStatusByAddr    ${votedAddress02}    #查看某地址的质押结果
    log    ${result}
    ${result}    getBalance    ${votedAddress02}    PTN
    log    ${result}
    Should Be Equal As Numbers    ${result}    10000
    ${result}    queryPledgeStatusByAddr    ${votedAddress03}    #查看某地址的质押结果
    log    ${result}
    ${result}    getBalance    ${votedAddress03}    PTN
    log    ${result}
    Should Be Equal As Numbers    ${result}    10000
    ${result}    mediatorListAll    #查看所有超级节点
    log    ${result}
    sleep    1
    ${mediatorAddress}    Set Variable    ${result[0]}
    ${result}    mediatorListVoteResults    #查看超级节点投票结果
    log    ${result}
    sleep    1
    ${result}    mediatorVote    ${votedAddress01}    ${mediatorAddress}    #投票某超级节点    #5000DAO
    log    ${result}
    sleep    5
    ${result}    mediatorVote    ${votedAddress02}    ${mediatorAddress}    #投票某超级节点    #5000DAO
    log    ${result}
    sleep    5
    ${result}    mediatorVote    ${votedAddress03}    ${mediatorAddress}    #投票某超级节点    #5000DAO
    log    ${result}
    sleep    5
    ${result}    mediatorGetVoted    ${votedAddress01}    #查看该节点所投票的情况
    log    ${result}
    ${result}    pledgeDeposit    ${votedAddress01}    100    #质押PTN    #101
    log    ${result}
    sleep    5
    ${result}    pledgeDeposit    ${votedAddress02}    100    #质押PTN    #101
    log    ${result}
    sleep    5
    ${result}    pledgeDeposit    ${votedAddress03}    100    #质押PTN    #101
    log    ${result}
    sleep    5
    ${result}    queryPledgeStatusByAddr    ${votedAddress01}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    NewDepositAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100
    ${result}    QueryPledgeHistoryByAddr    ${votedAddress01}
    log    ${result}
    ${result}    getBalance    ${votedAddress01}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress02}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    NewDepositAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100
    ${result}    getBalance    ${votedAddress02}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress03}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    NewDepositAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100
    ${result}    getBalance    ${votedAddress03}    PTN
    log    ${result}
    ${result}    mediatorListVoteResults    #查看超级节点投票结果
    log    ${result}
    ${amount}    Get From Dictionary    ${result}    ${mediatorAddress}
    sleep    1
    ${result}    queryPledgeList    #查看整个网络所有质押情况
    log    ${result}
    sleep    1
    ${time}    Get Time
    ${date}    Get Substring    ${time}    0    10
    log    ${date}
    ${yyyy}    ${mm}    ${dd} =    Get Time    year,month,day
    ${date}    Catenate    SEPARATOR=    ${yyyy}    ${mm}    ${dd}
    ${result}    QueryPledgeListByDate    ${date}
    log    ${result}
    sleep    1
    ${result}    QueryAllPledgeHistory
    log    ${result}
    sleep    1
    log    tiaojiaxinzhiya
    ${result}    isFinishAllocated
    log    ${result}
    Should Be Equal As Strings    ${result}    false
    sleep    5
    ${result}    HandlePledgeReward    ${votedAddress01}    #1
    log    ${result}
    sleep    5
    ${result}    isFinishAllocated
    log    ${result}
    Should Be Equal As Strings    ${result}    true
    sleep    5
    ${result}    queryPledgeList    #查看整个网络所有质押情况
    log    ${result}
    sleep    1
    ${result}    QueryPledgeListByDate    ${date}
    log    ${result}
    sleep    1
    ${result}    QueryAllPledgeHistory
    log    ${result}
    sleep    1
    ${result}    queryPledgeStatusByAddr    ${votedAddress01}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100
    ${result}    QueryPledgeHistoryByAddr    ${votedAddress01}
    log    ${result}
    ${result}    getBalance    ${votedAddress01}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress02}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100
    ${result}    getBalance    ${votedAddress02}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress03}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100
    ${result}    getBalance    ${votedAddress03}    PTN
    log    ${result}
    ${result}    mediatorListVoteResults    #查看超级节点投票结果
    log    ${result}
    ${amount}    Get From Dictionary    ${result}    ${mediatorAddress}
    sleep    1
    ${result}    mediatorVote    ${votedAddress04}    ${mediatorAddress}    #投票某超级节点    #5000DAO
    log    ${result}
    sleep    5
    ${result}    pledgeDeposit    ${votedAddress04}    100    #质押PTN    #101
    log    ${result}
    sleep    5
    ${result}    queryPledgeStatusByAddr    ${votedAddress04}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    NewDepositAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100
    ${result}    getBalance    ${votedAddress04}    PTN
    log    ${result}
    ${result}    mediatorVote    ${votedAddress05}    ${mediatorAddress}    #投票某超级节点    #5000DAO
    log    ${result}
    sleep    5
    ${result}    pledgeDeposit    ${votedAddress05}    100    #质押PTN    #101
    log    ${result}
    sleep    5
    ${result}    queryPledgeStatusByAddr    ${votedAddress05}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    NewDepositAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100
    ${result}    getBalance    ${votedAddress05}    PTN
    log    ${result}
    ${result}    getBalance    ${votedAddress06}    PTN
    log    ${result}
    sleep    1
    ${result}    pledgeDeposit    ${votedAddress06}    100    #质押PTN    #101
    log    ${result}
    sleep    5
    ${result}    mediatorVote    ${votedAddress06}    ${mediatorAddress}    #投票某超级节点    #5000DAO
    log    ${result}
    sleep    5
    ${result}    queryPledgeStatusByAddr    ${votedAddress06}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    NewDepositAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100
    ${result}    getBalance    ${votedAddress06}    PTN
    log    ${result}
    sleep    1
    ${result}    mediatorListVoteResults    #查看超级节点投票结果
    log    ${result}
    ${result}    pledgeDeposit    ${votedAddress01}    100    #质押PTN    #101
    log    ${result}
    sleep    5
    ${result}    queryPledgeStatusByAddr    ${votedAddress01}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    NewDepositAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100
    sleep    1
    ${result}    QueryPledgeHistoryByAddr    ${votedAddress01}
    log    ${result}
    ${result}    pledgeDeposit    ${votedAddress02}    100    #质押PTN    #101
    log    ${result}
    sleep    5
    ${result}    queryPledgeStatusByAddr    ${votedAddress02}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    NewDepositAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100
    sleep    1
    ${result}    mediatorListVoteResults    #查看超级节点投票结果
    log    ${result}
    ${result}    PledgeWithdraw    ${votedAddress01}    10000000000
    log    ${result}
    sleep    5
    ${result}    queryPledgeStatusByAddr    ${votedAddress01}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    WithdrawApplyAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100
    sleep    1
    ${result}    QueryPledgeHistoryByAddr    ${votedAddress01}
    log    ${result}
    ${result}    PledgeWithdraw    ${votedAddress02}    10000000000
    log    ${result}
    sleep    5
    ${result}    queryPledgeStatusByAddr    ${votedAddress02}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    WithdrawApplyAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100
    sleep    1
    ${result}    mediatorListVoteResults    #查看超级节点投票结果
    log    ${result}
    ${result}    isFinishAllocated
    log    ${result}
    sleep    5
    ${result}    HandlePledgeReward    ${foundationAddr}    #1
    log    ${result}
    sleep    5
    log    chuli2
    ${result}    isFinishAllocated
    log    ${result}
    sleep    5
    ${result}    HandlePledgeReward    ${foundationAddr}    #1
    log    ${result}
    sleep    5
    ${result}    isFinishAllocated
    log    ${result}
    sleep    5
    log    第一次分红
    ${result}    queryPledgeStatusByAddr    ${votedAddress01}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100.96248333
    ${result}    QueryPledgeHistoryByAddr    ${votedAddress01}
    log    ${result}
    ${result}    getBalance    ${votedAddress01}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress02}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100.96248333
    ${result}    getBalance    ${votedAddress02}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress03}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100.96248333
    ${result}    getBalance    ${votedAddress03}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress04}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100
    ${result}    getBalance    ${votedAddress04}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress05}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100
    ${result}    getBalance    ${votedAddress05}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress06}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100
    ${result}    getBalance    ${votedAddress06}    PTN
    log    ${result}
    sleep    1
    ${result}    queryPledgeList    #查看整个网络所有质押情况
    log    ${result}
    sleep    1
    ${result}    QueryAllPledgeHistory
    log    ${result}
    sleep    1
    ${result}    mediatorListVoteResults    #查看超级节点投票结果
    log    ${result}
    ${amount}    Get From Dictionary    ${result}    ${mediatorAddress}
    sleep    1
    log    第二次分红
    ${result}    isFinishAllocated
    log    ${result}
    sleep    5
    ${result}    HandlePledgeReward    ${foundationAddr}    #1
    log    ${result}
    sleep    5
    ${result}    isFinishAllocated
    log    ${result}
    sleep    5
    ${result}    HandlePledgeReward    ${foundationAddr}    #1
    log    ${result}
    sleep    5
    ${result}    isFinishAllocated
    log    ${result}
    sleep    5
    ${result}    queryPledgeStatusByAddr    ${votedAddress01}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    101.44602984
    ${result}    QueryPledgeHistoryByAddr    ${votedAddress01}
    log    ${result}
    ${result}    getBalance    ${votedAddress01}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress02}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    101.44602984
    ${result}    getBalance    ${votedAddress02}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress03}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    101.44602984
    ${result}    getBalance    ${votedAddress03}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress04}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100.47893682
    ${result}    getBalance    ${votedAddress04}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress05}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100.47893682
    ${result}    getBalance    ${votedAddress05}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress06}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100.47893682
    ${result}    getBalance    ${votedAddress06}    PTN
    log    ${result}
    ${result}    queryPledgeList    #查看整个网络所有质押情况
    log    ${result}
    sleep    1
    ${result}    QueryAllPledgeHistory
    log    ${result}
    sleep    1
    ${result}    PledgeWithdraw    ${votedAddress01}    10000000000
    log    ${result}
    sleep    5
    ${result}    queryPledgeStatusByAddr    ${votedAddress01}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    WithdrawApplyAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100
    sleep    1
    ${result}    QueryPledgeHistoryByAddr    ${votedAddress01}
    log    ${result}
    ${result}    PledgeWithdraw    ${votedAddress02}    10000000000
    log    ${result}
    sleep    5
    ${result}    queryPledgeStatusByAddr    ${votedAddress02}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    WithdrawApplyAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100
    sleep    1
    ${result}    PledgeWithdraw    ${votedAddress06}    10000000000
    log    ${result}
    sleep    5
    ${result}    queryPledgeStatusByAddr    ${votedAddress06}    #查看某地址的质押结果
    log    ${result}
    sleep    1
    ${result}    mediatorListVoteResults    #查看超级节点投票结果
    log    ${result}
    ${amount}    Get From Dictionary    ${result}    ${mediatorAddress}
    sleep    1
    log    第三次分红
    ${result}    isFinishAllocated
    log    ${result}
    sleep    1
    ${result}    HandlePledgeReward    ${foundationAddr}    #1
    log    ${result}
    sleep    5
    ${result}    isFinishAllocated
    log    ${result}
    sleep    2
    ${result}    HandlePledgeReward    ${foundationAddr}    #1
    log    ${result}
    sleep    5
    ${result}    isFinishAllocated
    log    ${result}
    sleep    5
    ${result}    queryPledgeStatusByAddr    ${votedAddress01}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    1.92957635
    ${result}    QueryPledgeHistoryByAddr    ${votedAddress01}
    log    ${result}
    ${result}    getBalance    ${votedAddress01}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress02}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    1.92957635
    ${result}    getBalance    ${votedAddress02}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress03}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    101.92957635
    ${result}    getBalance    ${votedAddress03}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress04}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100.95787364
    ${result}    getBalance    ${votedAddress04}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress05}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    100.95787364
    ${result}    getBalance    ${votedAddress05}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress06}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    0.95787364
    ${result}    getBalance    ${votedAddress06}    PTN
    log    ${result}
    ${result}    queryPledgeList    #查看整个网络所有质押情况
    log    ${result}
    sleep    1
    ${result}    QueryAllPledgeHistory
    log    ${result}
    sleep    1
    ${result}    mediatorListVoteResults    #查看超级节点投票结果
    log    ${result}
    ${amount}    Get From Dictionary    ${result}    ${mediatorAddress}
    sleep    1
    log    第四次分红
    ${result}    isFinishAllocated
    log    ${result}
    sleep    5
    ${result}    HandlePledgeReward    ${foundationAddr}    #1
    log    ${result}
    sleep    5
    log    chuli32
    ${result}    isFinishAllocated
    log    ${result}
    sleep    5
    ${result}    pledgeDeposit    ${votedAddress01}    100    #质押PTN    #101
    log    ${result}
    sleep    5
    ${result}    HandlePledgeReward    ${foundationAddr}    #1
    log    ${result}
    sleep    5
    ${result}    isFinishAllocated
    log    ${result}
    sleep    5
    ${result}    queryPledgeList    #查看整个网络所有质押情况
    log    ${result}
    sleep    1
    ${result}    QueryAllPledgeHistory
    log    ${result}
    sleep    1
    ${result}    mediatorListVoteResults    #查看超级节点投票结果
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress01}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    101.94762699
    ${result}    QueryPledgeHistoryByAddr    ${votedAddress01}
    log    ${result}
    ${result}    getBalance    ${votedAddress01}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress02}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    1.94762699
    ${result}    getBalance    ${votedAddress02}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress03}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    102.88309904
    ${result}    getBalance    ${votedAddress03}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress04}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    101.90230632
    ${result}    getBalance    ${votedAddress04}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress05}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    101.90230632
    ${result}    getBalance    ${votedAddress05}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress06}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    Should Be Equal As Strings    ${newDepositAmount}    0.96683428
    ${result}    getBalance    ${votedAddress06}    PTN
    log    ${result}
    ${amount}    getBalance    ${foundationAddr}    PTN
    log    ${amount}
    ${depositTwo}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM    PTN
    log    ${depositTwo}
    log    ${depositOne}
    #    Evaluate    ${depositTwo}-${depositOne}
    #    ${all}    311.5498
    log    第四次分红
    ${result}    isFinishAllocated
    log    ${result}
    sleep    5
    ${result}    HandlePledgeReward    ${foundationAddr}    #1
    log    ${result}
    sleep    5
    log    chuli32
    ${result}    isFinishAllocated
    log    ${result}
    sleep    5
    ${result}    HandlePledgeReward    ${foundationAddr}    #1
    log    ${result}
    sleep    5
    ${result}    isFinishAllocated
    log    ${result}
    sleep    5
    ${result}    queryPledgeStatusByAddr    ${votedAddress01}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    #    ${newDepositAmount}
    #    ${newDepositAmount}    1.94762699
    ${result}    QueryPledgeHistoryByAddr    ${votedAddress01}
    log    ${result}
    ${result}    getBalance    ${votedAddress01}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress02}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    #    ${newDepositAmount}    1.94762699
    ${result}    getBalance    ${votedAddress02}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress03}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    #    ${newDepositAmount}    102.88309904
    ${result}    getBalance    ${votedAddress03}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress04}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    #    ${newDepositAmount}    101.90230632
    ${result}    getBalance    ${votedAddress04}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress05}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    #    ${newDepositAmount}    101.90230632
    ${result}    getBalance    ${votedAddress05}    PTN
    log    ${result}
    ${result}    queryPledgeStatusByAddr    ${votedAddress06}    #查看某地址的质押结果
    log    ${result}
    ${resultJson}    To Json    ${result}
    ${newDepositAmount}    Get From Dictionary    ${resultJson}    PledgeAmount
    log    ${newDepositAmount}
    #    ${newDepositAmount}    0.96683428
    ${result}    getBalance    ${votedAddress06}    PTN
    log    ${result}
    ${amount}    getBalance    ${foundationAddr}    PTN
    log    ${amount}
    ${depositTwo}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM    PTN
    log    ${depositTwo}
    log    ${depositOne}
    ${result}    queryPledgeList    #查看整个网络所有质押情况
    log    ${result}
    sleep    1
    ${result}    QueryAllPledgeHistory
    log    ${result}
    sleep    1

Business_08
    [Documentation]    退出候选列表的两个Mediator继续交付保证金
    ...
    ...    mediator交付保证金再次进入超级节点和jury节点候选列表
    ${mDeposit}    getMediatorDepositWithAddr    ${mediatorAddr_01}
    log    ${mDeposit}
    ${amount}    getBalance    ${mediatorAddr_01}    PTN
    log    ${amount}
    Should Be Equal As Numbers    ${amount}    9996
    ${result}    mediatorPayToDepositContract    ${mediatorAddr_01}    50
    log    ${result}
    sleep    5
    ${mDeposit}    getMediatorDepositWithAddr    ${mediatorAddr_01}
    log    ${mDeposit}
    ${amount}    getBalance    ${mediatorAddr_01}    PTN
    log    ${amount}
    Should Be Equal As Numbers    ${amount}    9945
    ${depositOne}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM    PTN
    log    ${depositOne}
    ${addressMap3}    getListForMediatorCandidate
    log    ${addressMap3}
    Dictionary Should Contain Key    ${addressMap3}    ${mediatorAddr_01}
    ${resul}    getListForJuryCandidate
    Dictionary Should Contain Key    ${resul}    ${mediatorAddr_01}
    log    ${resul}
    ${mDeposit}    getMediatorDepositWithAddr    ${mediatorAddr_02}
    log    ${mDeposit}
    ${amount}    getBalance    ${mediatorAddr_02}    PTN
    log    ${amount}
    Should Be Equal As Numbers    ${amount}    9946
    ${result}    mediatorPayToDepositContract    ${mediatorAddr_02}    50
    log    ${result}
    sleep    5
    ${mDeposit}    getMediatorDepositWithAddr    ${mediatorAddr_02}
    log    ${mDeposit}
    ${amount}    getBalance    ${mediatorAddr_02}    PTN
    log    ${amount}
    Should Be Equal As Numbers    ${amount}    9895
    ${addressMap3}    getListForMediatorCandidate
    log    ${addressMap3}
    Dictionary Should Contain Key    ${addressMap3}    ${mediatorAddr_02}
    ${resul}    getListForJuryCandidate
    Dictionary Should Contain Key    ${resul}    ${mediatorAddr_02}
    log    ${resul}
    ${depositTwo}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM    PTN
    log    ${depositTwo}
    ${all}    Evaluate    ${depositTwo}-${depositOne}
    Should Be Equal As Numbers    ${all}    50
