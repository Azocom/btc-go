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
$dataAgora = date('Y-m-d H:i:s');

if (
        $password != 'oz48ClzUT2oq4xL6'
) {
        throw new \Exception("US-Erro");
}


try {
        $pdo = new PDO('mysql:host=' . $_CONFIG['HOST'] . ';dbname=' . $_CONFIG['BASE'], $_CONFIG['USER'], $_CONFIG['PASS']);
        $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

        $pdo->beginTransaction();

        if ($concluidoID != 0) {
                $sql = "
        UPDATE T_Blocos
        SET IC_Valid = 'S',
        hostname_valid = '{$miner_id}',
        DT_Valid = '{$dataAgora}'
        WHERE ID_T_Blocos = :ID_T_Blocos;
        ";
                $stm = $pdo->prepare($sql);
                $stm->execute([
                        'ID_T_Blocos' => $concluidoID
                ]);
        }

        $sql2 = "SELECT * FROM T_Blocos WHERE IC_Exec = 'N' AND  IC_Valid = 'N' ORDER BY ID_T_Blocos ASC LIMIT 1;";
        $stm = $pdo->prepare($sql2);
        $stm->execute();
        $resultado = $stm->fetch();

        $sql3 = "
        UPDATE T_Blocos
        SET IC_Exec = 'S',
        hostname = '{$miner_id}',
        DT_Inicio = '{$dataAgora}'
        WHERE ID_T_Blocos = :ID_T_Blocos;
        ";

        $stm = $pdo->prepare($sql3);
        $stm->execute([
                'ID_T_Blocos' => $resultado['ID_T_Blocos']
        ]);

        if (!$pdo->commit()) {
                $pdo->rollback();
                throw new \Exception("BT.U - Erro");
        }

        // $result['id'] = $resultado->ID_T_Blocos;
        $result['id'] = $resultado['ID_T_Blocos'];
        $result['min'] = $resultado['Inicio'];
        $result['max'] = $resultado['Fim'];
        $result['status'] = 200;
} catch (PDOException $e) {

        $result['status'] = 400;
        $result['miner_id'] = $miner_id;
        $result['error'] = $e->getMessage();
}


echo json_encode($result);
