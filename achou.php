<?php
header("Cache-Control: no-cache, no-store, must-revalidate"); // limpa o cache
header("Access-Control-Allow-Origin: *");
header("Access-Control-Allow-Methods: GET");
header("Access-Control-Allow-Headers: Origin,X-Requested-With,Content-Type, Accept");
header("Content-Type: application/json; charset=utf-8");

clearstatcache(); // limpa o cache

$_CONFIG['HOST'] = '34.95.167.202';
$_CONFIG['USER'] = 'acesso';
$_CONFIG['PASS'] = 'CFm7m9iP[nIzR5(';
$_CONFIG['BASE'] = 'diversos';

$miner_id = $_GET['u'];
$password = $_GET['p'];
$concluidoID = (int)$_GET['close_id'];
$dados = $_GET['dados'];
$dataAgora = date('Y-m-d H:i:s');

if (
        $password != '1E6NuFjCi27W5zoXg8TRdcSRq84zJeBW3k'
) {
        throw new \Exception("US-Erro");
}


try {
        $pdo = new PDO('mysql:host=' . $_CONFIG['HOST'] . ';dbname=' . $_CONFIG['BASE'], $_CONFIG['USER'], $_CONFIG['PASS']);
        $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

        if ($concluidoID != 0) {
                $sql = "
        UPDATE T_Blocos
        SET IC_Valid = 'S',
        hostname_achou = '{$miner_id}',
        DT_Valid = '{$dataAgora}',
        DS_Result = '{$dados}'
        WHERE ID_T_Blocos = :ID_T_Blocos;
        ";
                $stm = $pdo->prepare($sql);
                $stm->execute([
                        'ID_T_Blocos' => $concluidoID
                ]);
        }

        // $result['id'] = $resultado->ID_T_Blocos;
        $result['id'] = $concluidoID;
        $result['status'] = 200;
} catch (PDOException $e) {

        $result['status'] = 400;
        $result['miner_id'] = $miner_id;
        $result['error'] = $e->getMessage();
}


echo json_encode($result);
