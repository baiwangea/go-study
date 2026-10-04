<?php
class Animal {
    public function speak(): string {
        return "...";
    }
}

class Dog extends Animal {
    public function speak(): string {
        return "Woof";
    }
}

interface Loggable {
    public function log(): string;
}

class Cat extends Animal implements Loggable {
    public function speak(): string {
        return "Meow";
    }

    public function log(): string {
        return "cat spoke";
    }
}

$model = new Cat();
$a = $model->speak();
$b = $model->log();
var_dump($a, $b);
